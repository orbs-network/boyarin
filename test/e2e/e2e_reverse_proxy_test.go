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
	"os"
	"testing"
	"time"
)

func Test_UpdateReverseProxyWithSwarm(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipped on CI because of flakiness")
	}

	helpers.SkipUnlessSwarmIsEnabled(t)

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

	err = s.UpdateReverseProxy(context.Background(), &strelets.UpdateReverseProxyInput{
		chains, ip, adapter.SSLOptions{},
	})
	require.NoError(t, err)
	defer api.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	require.True(t, helpers.Eventually(1*time.Minute, func() bool {
		url := fmt.Sprintf("http://%s/vchains/%d/test", ip, chain.Id)
		fmt.Println(url)

		client := http.Client{
			Timeout: 2 * time.Second,
		}
		res, err := client.Get(url)
		if err != nil {
			fmt.Println("ERROR: could not access", url, ":", err)
			return false
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("ERROR: could not read", url, ":", err)
			return false
		}

		return res.StatusCode == 200 && string(body) == "success"
	}))
}

func Test_CreateReverseProxyWithSSL(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipped on CI because of flakiness")
	}

	helpers.SkipUnlessSwarmIsEnabled(t)

	port := 10099
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

	err = s.UpdateReverseProxy(context.Background(), &strelets.UpdateReverseProxyInput{
		chains, ip, adapter.SSLOptions{"./fixtures/cert.pem", "./fixtures/key.pem"},
	})
	require.NoError(t, err)
	defer api.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	require.True(t, helpers.Eventually(1*time.Minute, func() bool {
		url := fmt.Sprintf("https://%s:443/vchains/%d/test", ip, chain.Id)
		fmt.Println(url)

		client := http.Client{
			Timeout: 2 * time.Second,
		}
		_, err := client.Get(url)
		fmt.Println(err.Error())
		return err.Error() == fmt.Sprintf("Get %s: x509: cannot validate certificate for 10.4.12.224 because it doesn't contain any IP SANs", url)
	}))
}
