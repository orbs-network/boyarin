package helpers

import (
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

type ConfigFile int

const (
	Config ConfigFile = iota
	ConfigWithActiveVchains
	ConfigWithSingleChain
	ConfigWithSigner
)

func (conf ConfigFile) String() string {
	switch conf {
	case Config:
		return "config.json"
	case ConfigWithActiveVchains:
		return "configWithActiveVchains.json"
	case ConfigWithSingleChain:
		return "configWithSingleChain.json"
	case ConfigWithSigner:
		return "configWithSigner.json"
	default:
		panic(fmt.Sprintf("unknown config: %d", conf))
	}
}

func GetJSONConfig(t *testing.T, basePath string, conf ConfigFile) config.MutableNodeConfiguration {
	contents, err := ioutil.ReadFile(basePath + "/" + conf.String())
	require.NoError(t, err)
	source, err := config.NewStringConfigurationSource(string(contents), EthEndpoint()) // ethereum endpoint is optional
	require.NoError(t, err)
	return source
}
