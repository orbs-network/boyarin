package strelets

import (
	"encoding/json"
	"sort"
)

type FederationNode struct {
	Key  string
	IP   string
	Port int
}

func getNetworkConfigJSON(peers *PeersMap) []byte {
	jsonMap := make(map[string]interface{})

	var nodes []FederationNode
	for key, peer := range *peers {
		nodes = append(nodes, FederationNode{string(key), peer.IP, peer.Port})
	}

	// A workaround for tests because range does not preserve key order over iteration
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Key > nodes[j].Key
	})

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}
