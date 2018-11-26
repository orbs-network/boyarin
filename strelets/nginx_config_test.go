package strelets

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNginxConfig(t *testing.T) {
	chains := []*VirtualChain{
		{
			Id:       42,
			HttpPort: 8081,
		},
	}

	require.EqualValues(t, `server { listen 80; location /42/ { proxy_pass http://192.168.0.1:8081/; } }`, getNginxConfig(chains, "192.168.0.1"))
}
