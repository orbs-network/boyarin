package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

func main() {
	orchestratorPtr := flag.String("orchestrator", "docker", "docker|swarm")
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	flag.Parse()

	var orchestrator adapter.Orchestrator
	var err error

	switch *orchestratorPtr {
	case "docker":
		orchestrator, err = adapter.NewDockerAPI()
	case "swarm":
		orchestrator, err = adapter.NewDockerSwarm()
	default:
		err = fmt.Errorf("could not recognize orchestrator: %s", *orchestratorPtr)
	}

	if err != nil {
		panic(err)
	}

	config, err := boyar.NewUrlConfigurationSource(*configUrlPtr)
	if err != nil {
		panic(err)
	}

	s := strelets.NewStrelets("_tmp", orchestrator)
	b := boyar.NewBoyar(s, config, *keyPairConfigPathPtr)

	if err := b.ProvisionVirtualChains(context.Background()); err != nil {
		panic(err)
	}
}
