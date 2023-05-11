module github.com/orbs-network/boyarin

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/orbs-network/govnr v0.2.0
	github.com/orbs-network/scribe v0.2.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/shirou/gopsutil v2.20.9+incompatible
	github.com/stretchr/testify v1.4.0
	gotest.tools v2.2.0+incompatible // indirect
)

// Docker v19.03.5
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
