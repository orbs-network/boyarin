package adapter

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNginxContainerConfig(t *testing.T) {
	config := getNginxContainerConfig("nginx:1.13.1", "/tmp/config")
	result, _ := json.Marshal(config)

	expectedConfig := `{"ExposedPorts":{"80/tcp":{}},"HostConfig":{"Binds":["/tmp/config:/etc/nginx/conf.d"],"PortBindings":{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"80"}]}},"Image":"nginx:1.13.1"}`
	require.JSONEq(t, expectedConfig, string(result))
}
