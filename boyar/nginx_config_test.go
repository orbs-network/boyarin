package boyar

import (
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNginxConfig(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)

	//fmt.Println(getNginxConfig(cfg))

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / {
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location @error403 { return 403 '{"Status":"Forbidden","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
set $vc42 chain-42;
location ~ ^/vchains/42/logs/(.*) {
	alias /opt/orbs/logs/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/logs$ {
	alias /opt/orbs/logs/chain-42/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status/(.*) {
	alias /opt/orbs/status/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/chain-42/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location ~ ^/services/boyar/logs/(.*) {
	alias /opt/orbs/logs/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs/(.*) {
	alias /opt/orbs/logs/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs/(.*) {
	alias /opt/orbs/logs/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithDisabledChains(t *testing.T) {
	cfg := getJSONConfig(t, Config)

	fmt.Println(getNginxConfig(cfg))

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / {
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location @error403 { return 403 '{"Status":"Forbidden","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
set $vc42 chain-42;
location ~ ^/vchains/42/logs/(.*) {
	alias /opt/orbs/logs/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/logs$ {
	alias /opt/orbs/logs/chain-42/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status/(.*) {
	alias /opt/orbs/status/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/chain-42/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
set $vc1991 chain-1991;
location ~ ^/vchains/1991/logs/(.*) {
	alias /opt/orbs/logs/chain-1991/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/1991/logs$ {
	alias /opt/orbs/logs/chain-1991/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/1991/status/(.*) {
	alias /opt/orbs/status/chain-1991/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/1991/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/chain-1991/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/1991(/?)(.*) {
	proxy_pass http://$vc1991:8080/$2;
	error_page 502 = @error502;
}
location ~ ^/services/boyar/logs/(.*) {
	alias /opt/orbs/logs/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/logs/(.*) {
	alias /opt/orbs/logs/service-name/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/logs$ {
	alias /opt/orbs/logs/service-name/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/service-name/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/service-name/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs/(.*) {
	alias /opt/orbs/logs/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
}`,
		getNginxConfig(cfg))
}

func Test_getNginxConfigWithSSL(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetSSLOptions(adapter.SSLOptions{
		"fake-cert-path", "fake-key-path",
	})

	//fmt.Println(getNginxConfig(cfg))

	require.EqualValues(t, `server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
location ~^/$ { return 200 '{"Status":"OK","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location / {
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location @error403 { return 403 '{"Status":"Forbidden","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
set $vc42 chain-42;
location ~ ^/vchains/42/logs/(.*) {
	alias /opt/orbs/logs/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/logs$ {
	alias /opt/orbs/logs/chain-42/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status/(.*) {
	alias /opt/orbs/status/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/chain-42/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location ~ ^/services/boyar/logs/(.*) {
	alias /opt/orbs/logs/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs/(.*) {
	alias /opt/orbs/logs/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs/(.*) {
	alias /opt/orbs/logs/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
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
location / {
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location @error403 { return 403 '{"Status":"Forbidden","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error404 { return 404 '{"Status":"Not found","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
location @error502 { return 502 '{"Status":"Bad gateway","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":{"Semantic":"","Commit":""}}}}'; }
set $vc42 chain-42;
location ~ ^/vchains/42/logs/(.*) {
	alias /opt/orbs/logs/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/logs$ {
	alias /opt/orbs/logs/chain-42/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status/(.*) {
	alias /opt/orbs/status/chain-42/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/chain-42/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/42(/?)(.*) {
	proxy_pass http://$vc42:8080/$2;
	error_page 502 = @error502;
}
location ~ ^/services/boyar/logs/(.*) {
	alias /opt/orbs/logs/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/boyar/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs/(.*) {
	alias /opt/orbs/logs/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/management-service/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs/(.*) {
	alias /opt/orbs/logs/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "(GET|POST)") {
      add_header "Access-Control-Allow-Origin"  *;
    }

    # Preflight requests
    if ($request_method = OPTIONS ) {
      add_header "Access-Control-Allow-Origin"  *;
      add_header "Access-Control-Allow-Methods" "GET, POST, OPTIONS, HEAD";
      add_header "Access-Control-Allow-Headers" "Authorization, Origin, X-Requested-With, Content-Type, Accept";
      return 200;
    }

    # CORS end

	alias /opt/orbs/status/signer/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
}`,
		getNginxConfig(cfg))
}
