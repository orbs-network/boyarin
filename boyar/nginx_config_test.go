package boyar

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

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; error_page 502 = @error502; }
}`,
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

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; error_page 502 = @error502; }
}`,
		getNginxConfig(chains, "192.168.0.1", false))
}

func Test_getNginxConfigWithSSL(t *testing.T) {
	chains := []*VirtualChain{
		{
			Id:       42,
			HttpPort: 8081,
		},
	}

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; error_page 502 = @error502; }
}
server {
listen 443;
ssl on;
ssl_certificate /var/run/secrets/ssl-cert;
ssl_certificate_key /var/run/secrets/ssl-key;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://192.168.0.1:8081/; error_page 502 = @error502; }
}`,
		getNginxConfig(chains, "192.168.0.1", true))
}
