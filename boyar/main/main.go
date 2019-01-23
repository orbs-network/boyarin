package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"os"
	"time"
)

func build(keyPairConfigPath string, configUrl string, prevConfigHash string) (configHash string, err error) {
	config, err := boyar.NewUrlConfigurationSource(configUrl)
	if err != nil {
		return
	}
	configHash = config.Hash()
	if configHash == prevConfigHash {
		return
	}

	orchestrator, err := adapter.NewDockerSwarm(config.OrchestratorOptions())
	if err != nil {
		return
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := boyar.NewBoyar(s, config, keyPairConfigPath)

	if err = b.ProvisionVirtualChains(context.Background()); err != nil {
		return
	}

	if err = b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return
	}

	return
}

func main() {
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Uint("polling-interval", 60, "how often to poll for configuration in daemon mode (in seconds)")

	flag.Parse()

	if *daemonizePtr {
		successfulConfigHash := ""

		for true {
			if configHash, err := build(*keyPairConfigPathPtr, *configUrlPtr, successfulConfigHash); err != nil {
				fmt.Println("ERROR:", err)
				fmt.Println("Latest successful configuration", successfulConfigHash)
			} else {
				fmt.Println("Successfully updated to latest configuration:", configHash)
				successfulConfigHash = configHash
			}
			<-time.After(time.Duration(*pollingIntervalPtr) * time.Second)
		}
	} else {
		if _, err := build(*keyPairConfigPathPtr, *configUrlPtr, ""); err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}

}
