package config

func NewStringConfigurationSource(input string, ethereumEndpoint string, keyConfigPath string) (MutableNodeConfiguration, error) {
	return parseStringConfig(input, ethereumEndpoint, keyConfigPath)
}
