package strelets

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type VirtualChainId uint32

type VirtualChain struct {
	Id           VirtualChainId
	HttpPort     int
	GossipPort   int
	DockerConfig DockerConfig
	Config       map[string]interface{}
	Disabled     bool
}

func GetVcidFromServiceName(serviceName string) int64 {
	tokens := strings.Split(serviceName, "-")
	if len(tokens) < 2 {
		return -1
	}

	result, err := strconv.ParseInt(tokens[len(tokens)-2], 10, 0)
	if err != nil {
		return -1
	}

	return result
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
