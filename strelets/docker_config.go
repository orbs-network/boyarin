package strelets

type Resource struct {
	Memory int64
	CPUs   float64
}

// In Gb with defaults of 100 and 2
type DockerVolumes struct {
	Blocks int
	Logs   int
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
	Volumes             DockerVolumes
}

func (c *DockerConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}
