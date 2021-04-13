#!/bin/sh

if [[ -z "$(ls -A ~/data)" ]];
then
    cp ~/config.json ~/data/config.json
    cp ~/media/* ~/data
else
    echo "~/data is not empty"
fi

./app "$@"