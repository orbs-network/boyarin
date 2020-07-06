package boyar

import (
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNginxConfig(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location ~ ^/boyar/logs {
	alias /opt/orbs/logs/boyar/current;
	access_log off;
}
set $vc42 chain-42;
location ~ ^/vchains/42/logs {
	alias /opt/orbs/logs/chain-42/current;
	access_log off;
}
location ~ ^/vchains/42/status {
	alias /opt/orbs/status/chain-42/status.json;
	access_log off;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location /services/management-service/logs {
	alias /opt/orbs/logs/management-service/current;
}
location /services/management-service/status {
	alias /opt/orbs/status/management-service/status.json;
}
location /services/signer/logs {
	alias /opt/orbs/logs/signer/current;
}
location /services/signer/status {
	alias /opt/orbs/status/signer/status.json;
}
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithDisabledChains(t *testing.T) {
	cfg := getJSONConfig(t, Config)

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location ~ ^/boyar/logs {
	alias /opt/orbs/logs/boyar/current;
	access_log off;
}
set $vc42 chain-42;
location ~ ^/vchains/42/logs {
	alias /opt/orbs/logs/chain-42/current;
	access_log off;
}
location ~ ^/vchains/42/status {
	alias /opt/orbs/status/chain-42/status.json;
	access_log off;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
set $vc1991 chain-1991;
location ~ ^/vchains/1991/logs {
	alias /opt/orbs/logs/chain-1991/current;
	access_log off;
}
location ~ ^/vchains/1991/status {
	alias /opt/orbs/status/chain-1991/status.json;
	access_log off;
}
location ~ ^/vchains/1991(/?)(.*) {
	proxy_pass http://$vc1991:8080/$2;
	error_page 502 = @error502;
}
location /services/custom-service/logs {
	alias /opt/orbs/logs/custom-service/current;
}
location /services/custom-service/status {
	alias /opt/orbs/status/custom-service/status.json;
}
location /services/signer/logs {
	alias /opt/orbs/logs/signer/current;
}
location /services/signer/status {
	alias /opt/orbs/status/signer/status.json;
}
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithSSL(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetSSLOptions(adapter.SSLOptions{
		"fake-cert-path", "fake-key-path",
	})

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location ~ ^/boyar/logs {
	alias /opt/orbs/logs/boyar/current;
	access_log off;
}
set $vc42 chain-42;
location ~ ^/vchains/42/logs {
	alias /opt/orbs/logs/chain-42/current;
	access_log off;
}
location ~ ^/vchains/42/status {
	alias /opt/orbs/status/chain-42/status.json;
	access_log off;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location /services/management-service/logs {
	alias /opt/orbs/logs/management-service/current;
}
location /services/management-service/status {
	alias /opt/orbs/status/management-service/status.json;
}
location /services/signer/logs {
	alias /opt/orbs/logs/signer/current;
}
location /services/signer/status {
	alias /opt/orbs/status/signer/status.json;
}
}
server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 443 ssl;
ssl_certificate /var/run/secrets/ssl-cert;
ssl_certificate_key /var/run/secrets/ssl-key;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location ~ ^/boyar/logs {
	alias /opt/orbs/logs/boyar/current;
	access_log off;
}
set $vc42 chain-42;
location ~ ^/vchains/42/logs {
	alias /opt/orbs/logs/chain-42/current;
	access_log off;
}
location ~ ^/vchains/42/status {
	alias /opt/orbs/status/chain-42/status.json;
	access_log off;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location /services/management-service/logs {
	alias /opt/orbs/logs/management-service/current;
}
location /services/management-service/status {
	alias /opt/orbs/status/management-service/status.json;
}
location /services/signer/logs {
	alias /opt/orbs/logs/signer/current;
}
location /services/signer/status {
	alias /opt/orbs/status/signer/status.json;
}
}`,
		getNginxConfig(cfg))
}
