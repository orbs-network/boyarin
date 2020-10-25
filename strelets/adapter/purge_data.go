package adapter

import "context"

// Only works for EFS volumes because they are shared
func (d *dockerSwarmOrchestrator) PurgeServiceData(ctx context.Context, containerName string) error {
	return nil
}

func (d *dockerSwarmOrchestrator) PurgeVchainData(ctx context.Context, nodeAddress string, containerName string) error {
	return nil
}
