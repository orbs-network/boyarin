package strelets

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/version"
	"strings"
)

func getBasicNginxConfig(locations string) string {
	return fmt.Sprintf(`server {
listen 80;
%s
}`, locations)
}

func getSSLNginxConfig(locations string) string {
	return fmt.Sprintf(`server {
listen 443;
ssl on;
ssl_certificate /var/run/secrets/ssl-cert;
ssl_certificate_key /var/run/secrets/ssl-key;
%s
}`, locations)
}

type boyar struct {
	Version version.Version
}


type services struct {
	Boyar boyar
}


type defaultNginxResponse struct {
	Status string
	Description string
	Services services
}

func getDefaultNginxResponse(status string) string {
	raw, _ := json.Marshal(defaultNginxResponse{
		Status: status,
		Description: "ORBS blockchain node",
		Services: services{
			Boyar: boyar{
				Version: version.GetVersion(),
			},
		},
	})

	return string(raw)
}

func getNginxLocationDefinition(path string, code int) string {
	return fmt.Sprintf(`location %s { return %d '%%s'; }`, path, code)
}

func getNginxLocations(chains []*VirtualChain, ip string) string {
	locations := []string {
		fmt.Sprintf(getNginxLocationDefinition("~^/$", 200), getDefaultNginxResponse("OK")),
		`location / { error_page 404 = @error404; }`,
		fmt.Sprintf(getNginxLocationDefinition("@error404", 404), getDefaultNginxResponse("Not found")),
		fmt.Sprintf(getNginxLocationDefinition("@error502", 502), getDefaultNginxResponse("Bad gateway")),
	}

	for _, chain := range chains {
		if chain.Disabled {
			continue
		}
		locations = append(locations, fmt.Sprintf(`location /vchains/%d/ { proxy_pass http://%s:%d/; error_page 502 = @error502; }`, chain.Id, ip, chain.HttpPort))
	}

	return strings.Join(locations, "\n")
}

func getNginxConfig(chains []*VirtualChain, ip string, sslEnabled bool) string {
	locations := getNginxLocations(chains, ip)
	servers := []string{
		getBasicNginxConfig(locations),
	}

	if sslEnabled {
		servers = append(servers, getSSLNginxConfig(locations))
	}

	return strings.Join(servers, "\n")
}
