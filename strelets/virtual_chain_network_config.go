package strelets

import "encoding/json"

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

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}
