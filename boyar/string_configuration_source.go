package boyar

func NewStringConfigurationSource(input string) (NodeConfiguration, error) {
	return parseStringConfig(input)
}
