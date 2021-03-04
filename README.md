# RSGE-Grafana

[![GitlabCI](https://gitlab.com/schoentoon/rsge-grafana/badges/master/pipeline.svg)](https://gitlab.com/schoentoon/rsge-grafana)

This data source plugin will allow plotting of RuneScape Grand Exchange prices directly in Grafana.

## Installation

In the case of a grafana installation through docker, you would just have to add the following environment variables
```
GF_INSTALL_PLUGINS=https://schoentoon.gitlab.io/rsge-grafana/schoentoon-rsge-datasource-1.0.0.zip;rsge-datasource
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=schoentoon-rsge-datasource
```

Installation on other installation types will probably work fairly similar. Afterwards simply start grafana, login as admin and configure a new datasource with this plugin as the datatype.

### Frontend

1. Install dependencies
```BASH
yarn install
```

2. Build plugin in development mode or run in watch mode
```BASH
yarn dev
```
or
```BASH
yarn watch
```
3. Build plugin in production mode
```BASH
yarn build
```

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

```bash
go get -u github.com/grafana/grafana-plugin-sdk-go
```

2. Build backend plugin binaries for Linux, Windows and Darwin:
```BASH
mage -v
```

3. List all available Mage targets for additional commands:
```BASH
mage -l
```