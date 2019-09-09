package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlConfiguration_ParsesDownloadedConfigCorrectly(t *testing.T) {
	config, err := NewUrlConfigurationSource("https://boyar-testnet-bootstrap.s3-us-west-2.amazonaws.com/boyar/config.json", "")

	require.NoError(t, err, "Expected to not fail")
	require.GreaterOrEqual(t, 5, len(config.Chains()), "Expecting to find the correct amount of chains in the downloaded config file")
}
