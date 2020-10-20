module github.com/orbs-network/boyarin

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/c9s/goprocinfo v0.0.0-20200311234719-5750cbd54a3b
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/orbs-network/govnr v0.2.0
	github.com/orbs-network/scribe v0.2.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/text v0.3.2 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

// Docker v19.03.5
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
