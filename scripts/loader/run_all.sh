function setup_perf() {
    install_package() {
        NODE=$1
        PKG=$2

        ssh $NODE "
            if ! dpkg -l | grep -qw $PKG; then
                echo 'Installing $PKG on $NODE...'
                sudo apt-get update && sudo apt-get install -y $PKG
            else
                echo '$PKG is already installed on $NODE' 
            fi
        "
    }

    internal_setup() {
        NODE=$1

        # Package name to install
        PKGS=(
            "linux-tools-common"
            "linux-tools-5.4.0-164-generic"
        )

        for pkg in "${PKGS[@]}"
        do
            install_package $NODE $pkg &
        done

        # Remove existing perf data
        ssh $1 "rm -f ~/perf.data"
    }

    for node in "$@"
    do
        internal_setup "$node" &
    done

    wait
}

# setup_perf "$@"

rate=(10 $(seq 100 100 2500))

Runtime=400
Average=1.0

for i in "${rate[@]}"; do
    echo " =====> Test with rate: $i pakcet per second =====>"
    sleep 5

    OutputDir="./data-${i}rps-${Average}ms"
    mkdir ./data
    bash generate_workload.sh $Runtime $i $Average
    scp -q *.csv $1:~/loader/data/traces/example
    bash run_loader.sh "$@"
    mv ./data $OutputDir

    sleep 15
done