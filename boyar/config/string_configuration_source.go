package config

func NewStringConfigurationSource(input string) (MutableNodeConfiguration, error) {
	return parseStringConfig(input)
}
