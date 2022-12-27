package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/filters"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"time"
)

func AssertServiceUp(t helpers.TestingT, ctx context.Context, serviceName string) {
	orchestrator, err := adapter.NewDockerSwarm(&adapter.OrchestratorOptions{}, helpers.DefaultTestLogger())
	require.NoError(t, err)

	statuses, err := orchestrator.GetStatus(ctx, 1*time.Second)
	require.NoError(t, err)

	ok := false
	for _, status := range statuses {
		if status.Name == serviceName && status.State == "started" {
			ok = true
			return
		}
	}

	require.True(t, ok, "service should be up")
}

func AssertVolumeExists(t helpers.TestingT, ctx context.Context, volume string) {
	client := helpers.DockerClient(t)

	res, err := client.VolumeList(ctx, filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: volume,
	}))
	require.NoError(t, err)

	require.Len(t, res.Volumes, 1)
	require.Equal(t, volume, res.Volumes[0].Name)
}

func AssertServiceDown(t helpers.TestingT, ctx context.Context, serviceName string) {
	orchestrator, err := adapter.NewDockerSwarm(&adapter.OrchestratorOptions{}, helpers.DefaultTestLogger())
	require.NoError(t, err)

	statuses, err := orchestrator.GetStatus(ctx, 1*time.Second)
	require.NoError(t, err)

	ok := true
	for _, status := range statuses {
		ok = ok && status.Name != serviceName
	}

	require.True(t, ok, "service should be down")
}

func AssertManagementServiceUp(t helpers.TestingT, port int) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/node/management", port))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	config := make(map[string]interface{})
	err = json.Unmarshal(body, &config)
	require.NoError(t, err)
}

func GetVChainMetrics(t helpers.TestingT, port int, vc VChainArgument) JsonMap {
	metrics := make(map[string]interface{})
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/metrics", port, vc.Id))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &metrics)
	require.NoError(t, err)
	return JsonMap{value: metrics}
}

func AssertServiceStatusExists(t helpers.TestingT, port int, service string) {
	status := make(map[string]interface{})
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/services/%s/status", port, service))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &status)
	require.NoError(t, err)
}

func AssertVchainStatusExists(t helpers.TestingT, port int, vc VChainArgument) {
	status := make(map[string]interface{})
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/status", port, vc.Id))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &status)
	require.NoError(t, err)
}

func AssertVchainLogsExist(t helpers.TestingT, port int, vc VChainArgument) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/logs", port, vc.Id))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}

func AssertServiceLogsExist(t helpers.TestingT, port int, service string) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/services/%s/logs", port, service))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}

func AssertVchainProfilingInfoExist(t helpers.TestingT, port int, vc VChainArgument) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/debug/pprof/goroutine?debug=2", port, vc.Id))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}
