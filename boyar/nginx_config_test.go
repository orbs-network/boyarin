package boyar

import (
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNginxConfig(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://cfc9e5-chain-42-stack:8080/; error_page 502 = @error502; }
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithDisabledChains(t *testing.T) {
	cfg := getJSONConfig(t, Config)

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://cfc9e5-chain-42-stack:8080/; error_page 502 = @error502; }
location /vchains/1991/ { proxy_pass http://cfc9e5-chain-1991-stack:8080/; error_page 502 = @error502; }
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithSSL(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetSSLOptions(adapter.SSLOptions{
		"fake-cert-path", "fake-key-path",
	})

	require.EqualValues(t, `server {
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location /vchains/42/ { proxy_pass http://cfc9e5-chain-42-stack:8080/; error_page 502 = @error502; }
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
location /vchains/42/ { proxy_pass http://cfc9e5-chain-42-stack:8080/; error_page 502 = @error502; }
}`,
		getNginxConfig(cfg))
}
