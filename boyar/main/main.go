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

func build(orchestratorName string, keyPairConfigPath string, configUrl string, prevConfigHash string) (configHash string, err error) {
	config, err := boyar.NewUrlConfigurationSource(configUrl)
	if err != nil {
		return
	}
	configHash = config.Hash()
	if configHash == prevConfigHash {
		return
	}

	orchestrator, err := getOrchestrator(orchestratorName, config.OrchestratorOptions())
	if err != nil {
		return
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets("_tmp", orchestrator)
	b := boyar.NewBoyar(s, config, keyPairConfigPath)

	if err := b.ProvisionVirtualChains(context.Background()); err != nil {
		return
	}

	if err := b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return
	}

	return
}

func getOrchestrator(orchestratorName string, options *strelets.OrchestratorOptions) (orchestrator adapter.Orchestrator, err error) {
	switch orchestratorName {
	case "docker":
		orchestrator, err = adapter.NewDockerAPI("_tmp")
	case "swarm":
		orchestrator, err = adapter.NewDockerSwarm(options)
	default:
		err = fmt.Errorf("could not recognize orchestrator: %s", orchestratorName)
	}

	return orchestrator, err
}

func main() {
	orchestratorPtr := flag.String("orchestrator", "docker", "docker|swarm")
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Uint("polling-interval", 60, "how often to poll for configuration in daemon mode (in seconds)")

	flag.Parse()

	if *daemonizePtr {
		successfulConfigHash := ""

		for true {
			if configHash, err := build(*orchestratorPtr, *keyPairConfigPathPtr, *configUrlPtr, successfulConfigHash); err != nil {
				fmt.Println("ERROR:", err)
				fmt.Println("Latest successful configuration", successfulConfigHash)
			} else {
				fmt.Println("Successfully updated to latest configuration:", configHash)
				successfulConfigHash = configHash
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
