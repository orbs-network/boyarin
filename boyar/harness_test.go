package boyar

import (
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const fakeKeyPairPath = "./test/fake-key-pair.json"
const configPath = "./config/test"

type configFile int

const (
	Config configFile = iota
	ConfigWithSigner
)

func (conf configFile) String() string {
	switch conf {
	case Config:
		return "config.json"
	case ConfigWithSigner:
		return "configWithSigner.json"
	default:
		panic(fmt.Sprintf("unknown config: %d", conf))
	}
}

func getJSONConfig(t *testing.T, conf configFile) config.MutableNodeConfiguration {
	contents, err := ioutil.ReadFile(configPath + "/" + conf.String())
	require.NoError(t, err)
	source, err := config.NewStringConfigurationSource(string(contents), "http://localhost:7545", fakeKeyPairPath, false) // ethereum endpoint is optional
	require.NoError(t, err)
	return source
}
