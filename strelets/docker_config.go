package strelets

// In Gb with defaults of 100 and 2
type DockerVolumes struct {
	Blocks int
	Logs   int
}

type DockerConfig struct {
	Image               string
	Tag                 string
	Pull                bool
	ContainerNamePrefix string
	Volumes             DockerVolumes
}

func (c *DockerConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}
