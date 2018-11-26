package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestStrelets_UpdateReverseProxyOnDocker(t *testing.T) {
	server := test.CreateHttpServer("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("success"))
	})
	server.Start()
	defer server.Shutdown()

	api, err := adapter.NewDockerAPI("_tmp")
	require.NoError(t, err)

	s := strelets.NewStrelets("_tmp", api)
	chain := chain(1)
	chain.HttpPort = server.Port()

	chains := []*strelets.VirtualChain{chain}
	ip := test.LocalIP()

	err = s.UpdateReverseProxy(context.Background(), chains, ip)
	require.NoError(t, err)
	defer api.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	require.True(t, test.Eventually(20*time.Second, func() bool {
		url := fmt.Sprintf("http://%s/%d/test", ip, chain.Id)
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

func TestStrelets_UpdateReverseProxyOnSwarm(t *testing.T) {
	server := test.CreateHttpServer("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("success"))
	})
	server.Start()
	defer server.Shutdown()

	api, err := adapter.NewDockerSwarm()
	require.NoError(t, err)

	s := strelets.NewStrelets("_tmp", api)
	chain := chain(1)
	chain.HttpPort = server.Port()

	chains := []*strelets.VirtualChain{chain}
	ip := test.LocalIP()

	err = s.UpdateReverseProxy(context.Background(), chains, ip)
	require.NoError(t, err)
	defer api.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	require.True(t, test.Eventually(20*time.Second, func() bool {
		url := fmt.Sprintf("http://%s/%d/test", ip, chain.Id)
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