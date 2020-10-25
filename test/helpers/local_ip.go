package helpers

import (
	"net"
	"os"
)

func LocalIP() string {
	if localIp := os.Getenv("LOCAL_IP"); localIp != "" {
		return localIp
	}

	ifaces, _ := net.Interfaces()

	for _, i := range ifaces {
		if addrs, err := i.Addrs(); err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil && ip.To4() != nil && ip.To4().String() != "127.0.0.1" {
					return ip.To4().String()
				}
			}
		}
	}

	return "127.0.0.1"
}
