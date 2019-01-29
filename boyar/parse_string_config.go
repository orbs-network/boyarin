package boyar

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/crypto"
)

func parseStringConfig(input string) (*stringConfigurationSource, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}

	return &stringConfigurationSource{
		value: value,
		hash:  crypto.CalculateHash([]byte(input)),
	}, nil
}
