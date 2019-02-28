package strelets

import (
	"encoding/json"
	"sort"
)

type FederationNode struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
}

func getNetworkConfigJSON(peers *PeersMap) []byte {
	jsonMap := make(map[string]interface{})

	var nodes []FederationNode
	for address, peer := range *peers {
		nodes = append(nodes, FederationNode{string(address), peer.IP, peer.Port})
	}

	// A workaround for tests because range does not preserve key order over iteration
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Address > nodes[j].Address
	})

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}
