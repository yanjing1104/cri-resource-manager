memtierd-setup

# Test that 700M+ is moved to numa node {0..3} from a process
MEME_CGROUP=e2e-meme MEME_BS=700M MEME_BWC=1 MEME_BWS=700M memtierd-meme-start

next-round() {
    local round_counter_var=$1
    local round_counter_val=${!1}
    local round_counter_max=$2
    local round_delay=$3
    if [[ "$round_counter_val" -ge "$round_counter_max" ]]; then
        return 1
    fi
    eval "$round_counter_var=$(($round_counter_val + 1))"
    sleep $round_delay
    return 0
}

# In general, it takes around 5 seconds to finish moving out
match-moved() {
    local moved_regexp=$1
    local target_numa_node=$2
    round_number=0
    while ! ( memtierd-command "stats -t move_pages -f csv | awk -F, \"{print \\\$6}\""; grep ${moved_regexp} <<< $COMMAND_OUTPUT); do
        echo "grep MOVED value to the target numa node ${target_numa_node} matching ${moved_regexp} not found"
        next-round round_number 10 1 || {
            error "timeout: memtierd did not expected amount of memory"
        }
    done
}

for target_numa_node in {0..3}
do
  MEMTIERD_YAML=""
  memtierd-start
  memtierd-command "pages -pid ${MEME_PID}"
  memtierd-command "mover -pages-to ${target_numa_node}"
  echo "waiting 700M+ to be moved to ${target_numa_node}"
  match-moved "${target_numa_node}\:0\.7[0-9][0-9]" ${target_numa_node}
  memtierd-stop
done

memtierd-meme-stop
