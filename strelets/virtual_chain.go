package strelets

import (
	"fmt"
	"io/ioutil"
)

type VirtualChainId uint32

type VirtualChain struct {
	Id           VirtualChainId
	HttpPort     int
	GossipPort   int
	DockerConfig DockerImageConfig
}

func (v *VirtualChain) getContainerName() string {
	return fmt.Sprintf("%s-chain-%d", v.DockerConfig.ContainerNamePrefix, v.Id)
}

func copyFile(source string, destination string) error {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("%s: %s", err, source)
	}

	return ioutil.WriteFile(destination, data, 0600)
}
