package boyar

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/version"
	"strings"
	"text/template"
)

func getDefaultNginxResponse(status string) string {
	rawVersion, _ := json.Marshal(version.GetVersion())
	return fmt.Sprintf(`{"Status":"%s","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":%s}}}`, status, string(rawVersion))
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
	{{- range .Chains }}
location /vchains/{{.Id}}/ { proxy_pass http://{{.ServiceId}}:8080/; error_page 502 = @error502; }
	{{- end }} {{- /* range .Chains */ -}}
{{- end -}} {{- /* define "locations" */ -}}
server {
listen 80;
{{ template "locations" .}}
}
{{- if .SslEnabled }}
server {
listen 443;
ssl on;
ssl_certificate /var/run/secrets/ssl-cert;
ssl_certificate_key /var/run/secrets/ssl-key;
{{template "locations" .}}
}
{{- end}} {{- /* if .SslEnabled */ -}}`))
	var transformedChains []struct {
		Id        config.VirtualChainId
		ServiceId string
	}

	for _, chain := range cfg.Chains() {
		if !chain.Disabled {
			transformedChains = append(transformedChains, struct {
				Id        config.VirtualChainId
				ServiceId string
			}{Id: chain.Id, ServiceId: adapter.GetServiceId(cfg.PrefixedContainerName(chain.GetContainerName()))})
		}
	}

	err := TplNginxConf.Execute(&sb, struct {
		Chains     interface{}
		SslEnabled bool
	}{
		Chains:     transformedChains,
		SslEnabled: cfg.SSLOptions().SSLCertificatePath != "" && cfg.SSLOptions().SSLPrivateKeyPath != "",
	})

	if err != nil {
		panic(err)
	}
	return sb.String()
}
