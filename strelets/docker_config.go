package strelets

type DockerImageConfig struct {
	Image               string
	Tag                 string
	Pull                bool
	ContainerNamePrefix string
}

func (c *DockerImageConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}
