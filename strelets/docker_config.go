package strelets

type Resource struct {
	Memory int64
	CPUs   float64
}

type DockerResources struct {
	Limits       Resource
	Reservations Resource
}

type DockerConfig struct {
	Image               string
	Tag                 string
	Pull                bool
	ContainerNamePrefix string
	Resources           DockerResources
}

func (c *DockerConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}
