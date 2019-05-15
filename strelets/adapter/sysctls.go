package adapter

func GetSysctls() map[string]string {
	sysctls := make(map[string]string)
	sysctls["net.core.somaxconn"] = "128000"

	return sysctls
}
