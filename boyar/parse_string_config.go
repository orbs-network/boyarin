package boyar

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func hash(input []byte) string {
	checksum := sha256.Sum256(input)
	return hex.EncodeToString(checksum[:])
}

func parseStringConfig(input string) (*stringConfigurationSource, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}

	return &stringConfigurationSource{
		value: value,
		hash:  hash([]byte(input)),
	}, nil
}
