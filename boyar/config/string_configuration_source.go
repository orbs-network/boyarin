package config

func NewStringConfigurationSource(input string, ethereumEndpoint string) (MutableNodeConfiguration, error) {
	return parseStringConfig(input, ethereumEndpoint)
}
