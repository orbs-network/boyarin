package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func Test_UpdateReverseProxyWithSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	port := 10080
	server := helpers.CreateHttpServer("/test", port, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("success"))
	})
	server.Start()
	defer server.Shutdown()

	api, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)

	s := strelets.NewStrelets(api)
	chain := chain(1)
	chain.HttpPort = server.Port()

	chains := []*strelets.VirtualChain{chain}
	ip := helpers.LocalIP()

	err = s.UpdateReverseProxy(context.Background(), chains, ip)
	require.NoError(t, err)
	defer api.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	require.True(t, helpers.Eventually(20*time.Second, func() bool {
		url := fmt.Sprintf("http://%s/vchains/%d/test", ip, chain.Id)
		fmt.Println(url)
		res, err := http.Get(url)
		if err != nil {
			return false
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return false
		}

		return res.StatusCode == 200 && string(body) == "success"
	}))
}
