package boyar

import (
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"strconv"
	"strings"
	"time"
)

func buildPeersMap(nodes []*strelets.FederationNode, gossipPort int) *strelets.PeersMap {
	peersMap := make(strelets.PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[strelets.NodeAddress(node.Address)] = &strelets.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}

func getVcidFromServiceName(serviceName string) int64 {
	tokens := strings.Split(serviceName, "-")
	result, err := strconv.ParseInt(tokens[len(tokens)-1], 10, 0)
	if err != nil {
		return -1
	}

	return result
}

func formatAsISO6801(t time.Time) string {
	return t.Format(time.RFC3339)
}

func aggregateErrors(errors []error) error {
	if errors == nil {
		return nil
	}

	var lines []string

	for _, err := range errors {
		lines = append(lines, err.Error())
	}

	return fmt.Errorf(strings.Join(lines, ", "))
}
