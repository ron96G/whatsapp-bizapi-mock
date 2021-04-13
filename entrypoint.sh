#!/bin/sh

if [[ -z "$(ls -A ~/data)" ]];
then
    cp ~/config.json ~/data/config.json
else
    echo "~/data is not empty"
fi

./app "$@"