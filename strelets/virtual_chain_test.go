package strelets

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getVcidFromServiceName(t *testing.T) {
	require.EqualValues(t, 42, GetVcidFromServiceName("orbs-network-chain-42-stack"))
	require.EqualValues(t, -1, GetVcidFromServiceName("orbs-network-signer-service-stack"))
	require.EqualValues(t, -1, GetVcidFromServiceName("http-api-reverse-proxy-stack"))
}
