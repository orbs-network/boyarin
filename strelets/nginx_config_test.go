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

	require.EqualValues(t, `server { listen 80; location / { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; add_header Content-Type application/json; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; } }`,
		getNginxConfig(chains, "192.168.0.1", false))
}

func Test_getNginxConfigWithDisabledChains(t *testing.T) {
	chains := []*VirtualChain{
		{
			Id:       1832,
			HttpPort: 8080,
			Disabled: true,
		},
		{
			Id:       42,
			HttpPort: 8081,
		},
	}

	require.EqualValues(t, `server { listen 80; location / { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; add_header Content-Type application/json; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; } }`,
		getNginxConfig(chains, "192.168.0.1", false))
}

func Test_getNginxConfigWithSSL(t *testing.T) {
	chains := []*VirtualChain{
		{
			Id:       42,
			HttpPort: 8081,
		},
	}

	require.EqualValues(t, `server { listen 80; location / { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; add_header Content-Type application/json; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; } }
server { listen 443; ssl on; ssl_certificate /var/run/secrets/ssl-cert; ssl_certificate_key /var/run/secrets/ssl-key; location / { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; add_header Content-Type application/json; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; } }`,
		getNginxConfig(chains, "192.168.0.1", true))
}
