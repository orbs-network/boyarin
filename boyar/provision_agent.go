package boyar

import (
	"context"
	"fmt"

	"github.com/orbs-network/boyarin/agent"
	"github.com/orbs-network/boyarin/utils"
)

func (b *boyar) ProvisionAgent(ctx context.Context) error {
	// get instance
	if b.agent == nil {
		b.agent = agent.GetInstance()
	} else {
		b.agent.Start(false)
	}

	// init agent config
	url := fmt.Sprintf("http://localhost:8080/node/0x%s/main.sh", b.config.NodeAddress())

	// init agent config
	config := agent.Config{
		IntervalMinute: 1,
		Url:            url,
	}
	agent.Init(&config)

	// start
	var errors []error
	b.agent.Start(true)

	// if _, err := b.orchestrator.GetOverlayNetwork(ctx, adapter.SHARED_SIGNER_NETWORK); err != nil {
	// 	return errors.Wrap(err, "failed creating network")
	// }

	// for serviceName, service := range b.config.Services() {
	// 	if err := b.provisionService(ctx, serviceName, service); err != nil {
	// 		errors = append(errors, err)
	// 	}
	// }

	return utils.AggregateErrors(errors)
}
