module github.com/orbs-network/boyarin

go 1.12

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/containerd/containerd v1.5.18 // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/orbs-network/govnr v0.2.0
	github.com/orbs-network/scribe v0.2.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/shirou/gopsutil v2.20.9+incompatible
	github.com/stretchr/testify v1.7.0
)

// Docker v19.03.5
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
