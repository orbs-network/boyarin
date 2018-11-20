package boyar

import "encoding/json"

func parseStringConfig(input string) (*stringConfigurationSource, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}

	return &stringConfigurationSource{
		value: value,
	}, nil
}
