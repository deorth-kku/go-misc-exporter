module github.com/deorth-kku/go-misc-exporter

go 1.26

require (
	github.com/anatol/smart.go v0.0.0-20260427185427-04c4679efd4e
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/deorth-kku/go-common v0.0.0-20260317104337-5638c0964789
	github.com/deorth-kku/ryzenadj-go v0.0.0-20241011071124-61aec1e40818
	github.com/dustin/go-humanize v1.0.1
	github.com/mt-inside/go-lmsensors v1.99.9-dev
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2
	github.com/robbiet480/go.nut v0.0.0-20240622015809-60e196249c53
	github.com/siku2/arigo v0.3.0
	github.com/stoewer/go-strcase v1.3.1
)

replace github.com/mt-inside/go-lmsensors v1.99.9-dev => github.com/deorth-kku/go-lmsensors v0.0.0-20260402032211-f65228fa8545

require (
	git.dolansoft.org/lorenz/go-zfs v0.0.0-20241011010404-ba106a1b6427
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/hub v1.0.2 // indirect
	github.com/cenkalti/rpc2 v1.0.5 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/sys v0.43.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
