#!/bin/bash

function start_loader() {
    LATENCY_FILE="~/loader/data/out/*.csv"

    echo "Start loader on master: $1"

    # Remove existing cvs
    ssh $1 "rm -f " $LATENCY_FILE

    ssh $1 'tmux has-session -t loader 2>/dev/null || tmux new-session -d -s loader; tmux attach-session -t loader'
    ssh $1 'tmux send -t loader "cd ~/loader && go run cmd/loader.go --config cmd/config.json" ENTER'

    wait
    
    echo "Fetch loader result from master: $1"
    scp -q $1:~/loader/data/out/*.csv ./data/latency.csv
    scp -q -r $1:~/loader/data/traces/example ./data/setup
}

function perf_ndoes() {
    DURATION=300        # Total duration to run the script
    SAMPLE_DURATION=20  # Duration of each perf sampling

    ITERATIONS=$((DURATION / SAMPLE_DURATION))

    internal_remote_perf_oneshot() {
        NODE=$1

        NODE_NAME=$(echo "$NODE" | cut -d'@' -f2 | cut -d'.' -f1)

        echo "Start to perf node: $NODE_NAME"

        REMOTE_PERF_DATA_PATH="~/perf.data"

        # Start perf on the remote server
        ssh $NODE "sudo timeout $SAMPLE_DURATION perf record -F 99 -a -g -o $REMOTE_PERF_DATA_PATH"

        LOCAL_PERF_DATA_DIR="./data/perf"
        mkdir $LOCAL_PERF_DATA_DIR
        LOCAL_PERF_DATA_PATH="$LOCAL_PERF_DATA_DIR/$NODE_NAME-perf"

        # Retrieve the perf data file
        ssh $NODE "sudo chown yihan $REMOTE_PERF_DATA_PATH"
        scp -q -r $NODE:$REMOTE_PERF_DATA_PATH $LOCAL_PERF_DATA_PATH
    }

    internal_remote_perf_periodic() {
        NODE=$1

        NODE_NAME=$(echo "$NODE" | cut -d'@' -f2 | cut -d'.' -f1)

        echo "Start to perf node: $NODE_NAME"

        ssh $NODE "git clone https://github.com/brendangregg/FlameGraph.git ~/FlameGraph"

        REMOTE_PERF_DATA_DIR="~/perf"
        FLAMEGRAPH_DIR="~/FlameGraph"
        ssh $NODE "
            if [ -d "$REMOTE_PERF_DATA_DIR" ]; then
                rm -rf $REMOTE_PERF_DATA_DIR && mkdir $REMOTE_PERF_DATA_DIR
            else
                mkdir $REMOTE_PERF_DATA_DIR
            fi
        "

        for ((i=0; i<ITERATIONS; i++)); do
            START_SEC=$((i * SAMPLE_DURATION))
            END_SEC=$(((i+1) * SAMPLE_DURATION))

            echo " > Start to perf "$START_SEC"s to "$END_SEC"s on node $NODE_NAME..."

            REMOTE_PERF_DATA_PATH="$REMOTE_PERF_DATA_DIR/"$START_SEC"s-"$END_SEC"s.data"

            # Start perf on the remote server
            ssh $NODE "sudo timeout $SAMPLE_DURATION perf record -F 99 -a -g -o $REMOTE_PERF_DATA_PATH"
        done

        wait

        ssh $NODE "pushd "${REMOTE_PERF_DATA_DIR}" > /dev/null && for file in *-*.data; do
            start_sec=\$(echo \$file | cut -d '-' -f 1)
            end_sec=\$(echo \$file | cut -d '-' -f 2 | cut -d '.' -f 1)
            flamegraph_name="$NODE_NAME"-\"\$start_sec-\$end_sec.svg\"
            sudo perf script -i \"\$file\" | "$FLAMEGRAPH_DIR"/stackcollapse-perf.pl | "$FLAMEGRAPH_DIR"/flamegraph.pl > \"\$flamegraph_name\"
            sudo rm \$file
        done && popd > /dev/null"

        LOCAL_PERF_DATA_DIR="./data/perf/$NODE_NAME-perf"

        # Retrieve the perf data file
        ssh $NODE "sudo chown yihan $REMOTE_PERF_DATA_DIR/*"
        scp -q -r $NODE:$REMOTE_PERF_DATA_DIR $LOCAL_PERF_DATA_DIR
    }

    LOCAL_PERF_DIR="./data/perf"
    if [ -d "$LOCAL_PERF_DIR" ]; then
        rm -rf $LOCAL_PERF_DIR && mkdir $LOCAL_PERF_DIR
    else
        mkdir $LOCAL_PERF_DIR
    fi
    
    for node in "$@"
    do
        # internal_remote_perf_oneshot "$node" &
        internal_remote_perf_periodic "$node" &
    done

    wait
}

MASTER="$1"

perf_ndoes "$@" & start_loader $MASTER

wait