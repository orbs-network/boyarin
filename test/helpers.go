package test

import (
	"net"
	"time"
)

func LocalIP() string {
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

func PublicKeys() []string {
	return []string{
		"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"92d469d7c004cc0b24a192d9457836bf38effa27536627ef60718b00b0f33152",
		"a899b318e65915aa2de02841eeb72fe51fddad96014b73800ca788a547f8cce0",
	}
}

const eventuallyIterations = 25

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if f() {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}