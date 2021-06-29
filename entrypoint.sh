#!/bin/sh
set -e

if [[ -z "$(ls -A ~/data)" ]];
then
    cp ~/config.json ~/data/config.json
    cp ~/media/* ~/data
else
    echo "~/data is not empty"
fi

pid=0

trap_handler() {
    echo "trapped signal"
    if [ $pid -ne 0 ]; then
        kill -SIGTERM "$pid"
        wait "$pid"
    fi
    exit 143; # 128 + 15 -- SIGTERM
}

trap 'trap_handler' SIGINT SIGTERM

./app "$@" &
pid="$!"

echo "App running with pid ${pid}"
wait ${pid}
