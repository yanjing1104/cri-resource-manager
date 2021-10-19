// Copyright 2021 Intel Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memtier

import (
	"fmt"
	"strings"
)

func NewAddrRange(startAddr, stopAddr uint64) *AddrRange {
	if stopAddr < startAddr {
		startAddr, stopAddr = stopAddr, startAddr
	}
	return &AddrRange{addr: startAddr, length: (stopAddr - startAddr) / uint64(constPagesize)}
}

func (r *AddrRange) Length() uint64 {
	return r.length
}

func (ar *AddrRanges) Pid() int {
	return ar.pid
}

func (ar *AddrRanges) Ranges() []AddrRange {
	return ar.addrs
}

func (ar *AddrRanges) String() string {
	rs := []string{}
	for _, r := range ar.addrs {
		rs = append(rs, r.String())
	}
	s := fmt.Sprintf("AddrRanges{pid=%d ranges=%s}",
		ar.pid, strings.Join(rs, ","))
	return s
}

// Flatten returns AddrRanges where each item includes only one address range
func (ar *AddrRanges) Flatten() []*AddrRanges {
	rv := []*AddrRanges{}
	for _, r := range ar.addrs {
		newAr := &AddrRanges{
			pid:   ar.pid,
			addrs: []AddrRange{r},
		}
		rv = append(rv, newAr)
	}
	return rv
}

func (ar *AddrRanges) Filter(accept func(ar AddrRange) bool) *AddrRanges {
	newAr := &AddrRanges{
		pid:   ar.pid,
		addrs: []AddrRange{},
	}
	for _, r := range ar.addrs {
		if accept(r) {
			newAr.addrs = append(newAr.addrs, r)
		}
	}
	return newAr
}

func (ar *AddrRanges) SplitLength(maxLength uint64) *AddrRanges {
	newAr := &AddrRanges{
		pid:   ar.pid,
		addrs: make([]AddrRange, 0, len(ar.addrs)),
	}
	for _, r := range ar.addrs {
		addr := r.addr
		length := r.length
		for length > maxLength {
			newAr.addrs = append(newAr.addrs, AddrRange{addr, maxLength})
			length -= maxLength
			addr += maxLength
		}
		if length > 0 {
			newAr.addrs = append(newAr.addrs, AddrRange{addr, length})
		}
	}
	return newAr
}

func (r AddrRange) String() string {
	return fmt.Sprintf("%x(%d)", r.addr, r.length)
}

// PagesMatching returns pages with selected pagetable attributes
func (ar *AddrRanges) PagesMatching(pageAttributes uint64) (*Pages, error) {
	pages, err := procPagemap(ar.pid, ar.addrs, pageAttributes)
	if err != nil {
		return nil, err
	}
	return &Pages{pid: ar.pid, pages: pages}, nil
}

func (ar *AddrRanges) Intersection(intRanges []AddrRange) {
	newAddrs := []AddrRange{}
	for _, oldRange := range ar.addrs {
		for _, cutRange := range intRanges {
			start := oldRange.addr
			stop := oldRange.addr + oldRange.length*uint64(constPagesize)
			if cutRange.addr >= oldRange.addr &&
				cutRange.addr <= stop {
				if cutRange.addr > start {
					start = cutRange.addr
				}
				cutStop := cutRange.addr + cutRange.length*uint64(constPagesize)
				if cutStop < stop {
					stop = cutStop
				}
				newAddrs = append(newAddrs, *NewAddrRange(start, stop))
			}
		}
	}
	ar.addrs = newAddrs
}