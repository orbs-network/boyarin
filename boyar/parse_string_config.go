package boyar

import (
	"encoding/json"
)

func parseStringConfig(input string) (*nodeConfigurationContainer, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}

	return &nodeConfigurationContainer{
		value: value,
	}, nil
}
