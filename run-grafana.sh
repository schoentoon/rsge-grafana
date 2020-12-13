#!/bin/bash

docker run -it --rm -p 3000:3000 \
    -v /tmp/grafana:/var/lib/grafana \
    -v "$(pwd):/var/lib/grafana/plugins/rsge/" \
    -v "/home/schoentoon/.local/share/rscli/itemdb.ljson:/data/itemdb.ljson:ro" \
    -e "GF_LOG_LEVEL=debug" \
    -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=schoentoon-rsge" \
    grafana/grafana