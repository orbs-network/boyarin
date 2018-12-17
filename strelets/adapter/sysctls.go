package adapter

func getSysctls() map[string]string {
	sysctls := make(map[string]string)
	sysctls["net.core.somaxconn"] = "25000"

	return sysctls
}
