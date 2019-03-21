package strelets

import (
	"fmt"
	"strings"
)

func getNginxConfig(chains []*VirtualChain, ip string) string {
	var locations []string

	for _, chain := range chains {
		if chain.Disabled {
			continue
		}
		locations = append(locations, fmt.Sprintf(`location /vchains/%d/ { proxy_pass http://%s:%d/; }`, chain.Id, ip, chain.HttpPort))
	}

	return fmt.Sprintf(`server { listen 80; %s }`, strings.Join(locations, "\n"))
}
