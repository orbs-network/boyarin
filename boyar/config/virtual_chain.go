package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type VirtualChainId uint32

type VirtualChain struct {
	Service
	Id               VirtualChainId
	InternalHttpPort int // FIXME should be deprecated as vchain specific
}

type VirtualChainConfig struct {
	VirtualChain *VirtualChain
	Topology     []*FederationNode
	NodeAddress  NodeAddress

	KeyPairConfig []byte `json:"-"` // Prevents key leak via log
}

type Peer struct {
	IP   string
	Port int
}

func GetVcidFromServiceName(serviceName string) int64 {
	tokens := strings.Split(serviceName, "-")
	if len(tokens) < 2 {
		return -1
	}

	result, err := strconv.ParseInt(tokens[len(tokens)-1], 10, 0)
	if err != nil {
		return -1
	}

	return result
}

func (v *VirtualChain) GetContainerName() string {
	return fmt.Sprintf("chain-%d", v.Id)
}

func (c *VirtualChain) GetSerializedConfig() []byte {
	m := make(map[string]interface{})
	for k, v := range c.Config {
		m[k] = v
	}
	m["virtual-chain-id"] = c.Id

	rawJSON, _ := json.Marshal(m)
	return rawJSON
}

func (id VirtualChainId) String() string {
	return strconv.Itoa(int(id))
}

type FederationNode struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
}
