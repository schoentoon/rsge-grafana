#!/bin/bash

mkdir -p ./tmp/grafana

sudo chown 472:472 -R ./tmp

docker run -it --rm -p 3000:3000 \
    -v "$(pwd)/tmp/grafana:/var/lib/grafana" \
    -v "$(pwd):/var/lib/grafana/plugins/rsge/" \
    -v "/home/schoentoon/.local/share/rscli/itemdb.ljson:/data/itemdb.ljson:ro" \
    -e "GF_LOG_LEVEL=debug" \
    -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=schoentoon-rsge-datasource" \
    grafana/grafana