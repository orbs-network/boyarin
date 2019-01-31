package strelets

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type VirtualChainId uint32

type VirtualChain struct {
	Id           VirtualChainId
	HttpPort     int
	GossipPort   int
	DockerConfig DockerImageConfig `json:"DockerConfig"`
	Config       map[string]interface{}
}

func (v *VirtualChain) getContainerName() string {
	return fmt.Sprintf("%s-chain-%d", v.DockerConfig.ContainerNamePrefix, v.Id)
}

func (c *VirtualChain) getSerializedConfig() []byte {
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
