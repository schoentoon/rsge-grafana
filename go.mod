module github.com/grafana/simple-datasource-backend

go 1.14

require (
	github.com/grafana/grafana-plugin-sdk-go v0.65.0
	gitlab.com/schoentoon/rs-tools v0.0.0-00010101000000-000000000000
)

replace gitlab.com/schoentoon/rs-tools => ../rs-tools
