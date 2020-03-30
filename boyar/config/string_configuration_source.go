package config

func NewStringConfigurationSource(input string, ethereumEndpoint string, keyConfigPath string, withNamespace bool) (MutableNodeConfiguration, error) {
	return parseStringConfig(input, ethereumEndpoint, keyConfigPath, withNamespace)
}
