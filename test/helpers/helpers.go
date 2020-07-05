package helpers

import (
	"github.com/orbs-network/scribe/log"
	"net"
	"os"
	"testing"
	"time"
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

func NodeAddresses() []string {
	return []string{
		"a328846cd5b4979d68a8c58a9bdfeee657b34de7",
		"d27e2e7398e2582f63d0800330010b3e58952ff6",
		"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9",
		"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17",
	}
}

const eventuallyIterations = 50

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if f() {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}

func SkipUnlessSwarmIsEnabled(t *testing.T) {
	if os.Getenv("ENABLE_SWARM") != "true" {
		t.Skip("skipping test, docker swarm is disabled")
	}
}

func DefaultTestLogger() log.Logger {
	return log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))
}
