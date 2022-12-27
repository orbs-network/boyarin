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

location ~ ^/services/boyar/logs/(?<filename>.*) {
	alias /opt/orbs/logs/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/management-service/logs/(?<filename>.*) {
	alias /opt/orbs/logs/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/signer/logs/(?<filename>.*) {
	alias /opt/orbs/logs/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/boyar/logs/(?<filename>.*) {
	alias /opt/orbs/logs/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/service-name/logs/(?<filename>.*) {
	alias /opt/orbs/logs/service-name/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/logs$ {
	alias /opt/orbs/logs/service-name/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/service-name/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/service-name/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/signer/logs/(?<filename>.*) {
	alias /opt/orbs/logs/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

location ~ ^/services/boyar/logs/(?<filename>.*) {
	alias /opt/orbs/logs/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/management-service/logs/(?<filename>.*) {
	alias /opt/orbs/logs/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/signer/logs/(?<filename>.*) {
	alias /opt/orbs/logs/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

location ~ ^/services/boyar/logs/(?<filename>.*) {
	alias /opt/orbs/logs/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/logs$ {
	alias /opt/orbs/logs/boyar/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/boyar/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/boyar/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/management-service/logs/(?<filename>.*) {
	alias /opt/orbs/logs/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/logs$ {
	alias /opt/orbs/logs/management-service/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/management-service/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/management-service/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
location ~ ^/services/signer/logs/(?<filename>.*) {
	alias /opt/orbs/logs/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/logs$ {
	alias /opt/orbs/logs/signer/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status/(?<filename>.*) {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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

	alias /opt/orbs/status/signer/$filename;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/signer/status$ {

	# CORS start

    # Simple requests
    if ($request_method ~* "GET|POST") {
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
