package topology

type FederationNode struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
}
