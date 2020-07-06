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
	Port         int
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

func getNginxConfig(cfg config.NodeConfiguration) string {
	var sb strings.Builder
	var TplNginxConf = template.Must(template.New("").Funcs(template.FuncMap{
		"DefaultResponse": getDefaultNginxResponse,
	}).Parse(`{{ "" -}}
{{- define "locations" -}}
location ~^/$ { return 200 '{{DefaultResponse "OK"}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{{DefaultResponse "Not found"}}'; }
location @error502 { return 502 '{{DefaultResponse "Bad gateway"}}'; }
location ~ ^/boyar/logs {
	alias /opt/orbs/logs/boyar/current;
	access_log off;
}
{{- range .Chains }}
set $vc{{.Id}} {{.ServiceId}};
location ~ ^/vchains/{{.Id}}/logs {
	alias {{.LogsVolume}}/current;
	access_log off;
}
location ~ ^/vchains/{{.Id}}/status {
	alias {{.StatusVolume}}/status.json;
	access_log off;
}
location ~ ^/vchains/{{.Id}}(/?)(.*) {
	proxy_pass http://$vc{{.Id}}:{{.Port}}/$2;
	error_page 502 = @error502;
}
{{- end }} {{- /* range .Chains */ -}}
{{- range .Services }}
location /services/{{.ServiceId}}/logs {
	alias {{.LogsVolume}}/current;
}
location /services/{{.ServiceId}}/status {
	alias {{.StatusVolume}}/status.json;
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

	sort.Slice(services, func(i, j int) bool {
		return services[i].ServiceId < services[j].ServiceId
	})

	err := TplNginxConf.Execute(&sb, nginxTemplateParams{
		Chains:     transformedChains,
		Services:   services,
		SslEnabled: cfg.SSLOptions().SSLCertificatePath != "" && cfg.SSLOptions().SSLPrivateKeyPath != "",
	})

	if err != nil {
		panic(err)
	}
	return sb.String()
}
