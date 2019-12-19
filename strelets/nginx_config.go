package strelets

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/version"
	"strings"
	"text/template"
)

func getDefaultNginxResponse(status string) string {
	rawVersion, _ := json.Marshal(version.GetVersion())
	return fmt.Sprintf(`{"Status":"%s","Description":"ORBS blockchain node","Services":{"Boyar":{"Version":%s}}}`, status, string(rawVersion))
}

func getNginxConfig(chains []*VirtualChain, ip string, sslEnabled bool) string {
	var sb strings.Builder
	var TplNginxConf = template.Must(template.New("").Funcs(template.FuncMap{
		"DefaultResponse": getDefaultNginxResponse,
	}).Parse(`{{ "" -}}
{{- define "locations" -}}
location ~^/$ { return 200 '{{DefaultResponse "OK"}}'; }
location / { error_page 404 = @error404; }
location @error404 { return 404 '{{DefaultResponse "Not found"}}'; }
location @error502 { return 502 '{{DefaultResponse "Bad gateway"}}'; }
{{ $ip := .Ip -}}
	{{- range .Chains -}}
		{{- if not .Disabled -}}
location /vchains/{{.Id}}/ { proxy_pass http://{{$ip}}:{{.HttpPort}}/; error_page 502 = @error502; proxy_connect_timeout 75s;}
		{{- end -}} {{- /* if not .Disabled */ -}}
	{{- end -}} {{- /* range .Chains */ -}}
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
	err := TplNginxConf.Execute(&sb, struct {
		Chains     []*VirtualChain
		Ip         string
		SslEnabled bool
	}{
		Chains:     chains,
		Ip:         ip,
		SslEnabled: sslEnabled,
	})
	if err != nil {
		panic(err)
	}
	return sb.String()
}
