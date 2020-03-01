package boyar

import (
	"github.com/orbs-network/boyarin/boyar/config"
)

func getKeyConfigJson(config config.NodeConfiguration, addressOnly bool) []byte {
	keyConfig := config.KeyConfig()
	if keyConfig == nil {
		return []byte{}
	}
	return keyConfig.JSON(addressOnly)
}
