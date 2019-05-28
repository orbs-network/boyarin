package boyar

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getVcidFromServiceName(t *testing.T) {
	require.EqualValues(t, 42, getVcidFromServiceName("orbs-network-chain-42-stack"))
	require.EqualValues(t, -1, getVcidFromServiceName("orbs-network-signer-service-stack"))
	require.EqualValues(t, -1, getVcidFromServiceName("http-api-reverse-proxy-stack"))
}
