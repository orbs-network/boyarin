package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewUrlConfigurationSource(url string, ethereumEndpoint string, keyConfigPath string, withNamespace bool) (MutableNodeConfiguration, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not download configuration from source: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("management config url returned with status %s", resp.Status)
	}

	input, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration from source: %s", err)
	}

	return parseStringConfig(string(input), ethereumEndpoint, keyConfigPath, withNamespace)
}
