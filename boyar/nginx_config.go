package boyar

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/version"
	"sort"
	"strings"
	"text/template"
)

func getDefaultNginxResponse(status string) string {
	rawVersion, _ := json.Marshal(version.GetVersion())
	return fmt.Sprintf(`{"Status":"%s","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":%s}}}`, status, string(rawVersion))
}

type nginxTemplateChainParams struct {
	Id           config.VirtualChainId
	ServiceId    string
	Port         int // FIXME should be removed because it's always 8080
	LogsVolume   string
	StatusVolume string
}

type nginxTemplateServiceParams struct {
	ServiceId    string
	LogsVolume   string
	StatusVolume string
}

type nginxTemplateParams struct {
	Chains     []nginxTemplateChainParams
	Services   []nginxTemplateServiceParams
	SslEnabled bool
}

const BOYAR_SERVICE = "boyar"

func getNginxConfig(cfg config.NodeConfiguration) string {
	var sb strings.Builder
	var tplNginxConf = template.Must(template.New("").Funcs(template.FuncMap{
		"DefaultResponse": getDefaultNginxResponse, "CORS": getCORS,
	}).Parse(`{{ "" -}}
{{- define "locations" -}}
location ~^/$ { return 200 '{{DefaultResponse "OK"}}'; }
location / {
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location @error403 { return 403 '{{DefaultResponse "Forbidden"}}'; }
location @error404 { return 404 '{{DefaultResponse "Not found"}}'; }
location @error502 { return 502 '{{DefaultResponse "Bad gateway"}}'; }
{{- range .Chains }}
set $vc{{.Id}} {{.ServiceId}};
location ~ ^/vchains/{{.Id}}/logs/(.*) {
	alias {{.LogsVolume}}/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/{{.Id}}/logs$ {
	alias {{.LogsVolume}}/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/{{.Id}}/status/(.*) {
	alias {{.StatusVolume}}/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/{{.Id}}/status$ {
{{ CORS }}
	alias {{.StatusVolume}}/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/vchains/{{.Id}}(/?)(.*) {
	proxy_pass http://$vc{{.Id}}:{{.Port}}/$2;
	error_page 502 = @error502;
}
{{- end }} {{- /* range .Chains */ -}}
{{- range .Services }}
location ~ ^/services/{{.ServiceId}}/logs/(.*) {
	alias {{.LogsVolume}}/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/{{.ServiceId}}/logs$ {
	alias {{.LogsVolume}}/current;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/{{.ServiceId}}/status/(.*) {
{{ CORS }}
	alias {{.StatusVolume}}/$1;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
location ~ ^/services/{{.ServiceId}}/status$ {
{{ CORS }}
	alias {{.StatusVolume}}/status.json;
	error_page 404 = @error404;
	error_page 403 = @error403;
}
{{- end }}
{{- end -}} {{- /* define "locations" */ -}}
server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 80;
{{ template "locations" .}}
}
{{- if .SslEnabled }}
server {
access_log off;
error_log off;
resolver 127.0.0.11 ipv6=off;
listen 443 ssl;
ssl_certificate /var/run/secrets/ssl-cert;
ssl_certificate_key /var/run/secrets/ssl-key;
{{template "locations" .}}
}
{{- end}} {{- /* if .SslEnabled */ -}}`))
	var transformedChains []nginxTemplateChainParams

	for _, chain := range cfg.Chains() {
		if !chain.Disabled {
			containerName := cfg.NamespacedContainerName(chain.GetContainerName())
			transformedChains = append(transformedChains, nginxTemplateChainParams{
				Id:           chain.Id,
				ServiceId:    containerName,
				Port:         chain.InternalHttpPort,
				LogsVolume:   adapter.GetNginxLogsMountPath(chain.GetContainerName()),
				StatusVolume: adapter.GetNginxStatusMountPath(chain.GetContainerName()),
			})
		}
	}

	var services []nginxTemplateServiceParams
	for serviceName, _ := range cfg.Services() {
		services = append(services, nginxTemplateServiceParams{
			ServiceId:    serviceName,
			LogsVolume:   adapter.GetNginxLogsMountPath(serviceName),
			StatusVolume: adapter.GetNginxStatusMountPath(serviceName),
		})
	}

	// special case to pass boyar logs from the outside
	services = append(services, nginxTemplateServiceParams{
		ServiceId:    BOYAR_SERVICE,
		LogsVolume:   adapter.GetNginxLogsMountPath(BOYAR_SERVICE),
		StatusVolume: adapter.GetNginxStatusMountPath(BOYAR_SERVICE),
	})

	sort.Slice(services, func(i, j int) bool {
		return services[i].ServiceId < services[j].ServiceId
	})

	err := tplNginxConf.Execute(&sb, nginxTemplateParams{
		Chains:     transformedChains,
		Services:   services,
		SslEnabled: cfg.SSLOptions().SSLCertificatePath != "" && cfg.SSLOptions().SSLPrivateKeyPath != "",
	})

	if err != nil {
		panic(err)
	}
	return sb.String()
}

func getCORS() string {
	return `
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
`
}
