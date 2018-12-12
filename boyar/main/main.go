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

func build(orchestratorName string, keyPairConfigPath string, configUrl string, prevConfigHash string) (err error, configHash string) {
	var orchestrator adapter.Orchestrator

	switch orchestratorName {
	case "docker":
		orchestrator, err = adapter.NewDockerAPI("_tmp")
	case "swarm":
		orchestrator, err = adapter.NewDockerSwarm()
	default:
		err = fmt.Errorf("could not recognize orchestrator: %s", orchestratorName)
	}

	if err != nil {
		return err, ""
	}

	config, err := boyar.NewUrlConfigurationSource(configUrl)
	if err != nil {
		return err, ""
	}
	configHash = config.Hash()

	if configHash == prevConfigHash {
		return nil, configHash
	}

	s := strelets.NewStrelets("_tmp", orchestrator)
	b := boyar.NewBoyar(s, config, keyPairConfigPath)

	if err := b.ProvisionVirtualChains(context.Background()); err != nil {
		return err, configHash
	}

	if err := b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return err, configHash
	}

	return nil, configHash
}

func main() {
	orchestratorPtr := flag.String("orchestrator", "docker", "docker|swarm")
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Uint("polling-interval", 60, "how often to poll for configuration in daemon mode (in seconds)")

	flag.Parse()

	if *daemonizePtr {
		configHash := ""
		var err error

		for true {
			if err, configHash = build(*orchestratorPtr, *keyPairConfigPathPtr, *configUrlPtr, configHash); err != nil {
				fmt.Println("ERROR:", err)
			} else {
				fmt.Println("Successfully updated to latest configuration:", configHash)
			}
			<-time.After(time.Duration(*pollingIntervalPtr) * time.Second)
		}
	} else {
		if err, _ := build(*orchestratorPtr, *keyPairConfigPathPtr, *configUrlPtr, ""); err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}

}
