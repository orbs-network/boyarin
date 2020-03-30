package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getVcidFromServiceName(t *testing.T) {
	require.EqualValues(t, 42, GetVcidFromServiceName("chain-42"))
	require.EqualValues(t, -1, GetVcidFromServiceName("signer"))
	require.EqualValues(t, -1, GetVcidFromServiceName("http-api-reverse-proxy"))
}
