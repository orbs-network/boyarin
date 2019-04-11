package strelets

import (
	"fmt"
	"strings"
)

func getBasicNginxConfig(locations string) string {
	return fmt.Sprintf(`server { listen 80; %s }`, locations)
}

func getSSLNginxConfig(locations string) string {
	return fmt.Sprintf(`server { listen 443; ssl on; ssl_certificate /var/run/secrets/ssl-cert; ssl_certificate_key /var/run/secrets/ssl-key; %s }`, locations)
}

func getNginxLocations(chains []*VirtualChain, ip string) string {
	var locations []string

	for _, chain := range chains {
		if chain.Disabled {
			continue
		}
		locations = append(locations, fmt.Sprintf(`location /vchains/%d/ { proxy_pass http://%s:%d/; }`, chain.Id, ip, chain.HttpPort))
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
