package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"io/ioutil"
)

func (d *dockerSwarm) GetStatus(ctx context.Context) (results []*ContainerStatus, err error) {
	if tasks, err := d.client.TaskList(ctx, types.TaskListOptions{}); err != nil {
		return nil, fmt.Errorf("failed to retrieve task list: %s", err)
	} else {
		for _, task := range tasks {
			name, _ := d.getServiceName(ctx, task.ServiceID)
			logs, _ := d.getLogs(ctx, task.ServiceID)

			results = append(results, &ContainerStatus{
				Name:      name,
				State:     task.Status.Message,
				Error:     task.Status.Err,
				NodeID:    task.NodeID,
				CreatedAt: task.CreatedAt,
				Logs:      logs,
			})
		}
	}

	return
}

func (d *dockerSwarm) getServiceName(ctx context.Context, serviceID string) (string, error) {
	if specs, err := d.client.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			"id",
			serviceID,
		}),
	}); err != nil {
		return "", err
	} else if len(specs) == 0 {
		return "", fmt.Errorf("no such service")
	} else {
		return specs[0].Spec.Name, nil
	}
}

func (d *dockerSwarm) getLogs(ctx context.Context, serviceID string) (string, error) {
	io, err := d.client.ServiceLogs(ctx, serviceID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		return "", fmt.Errorf("could not retrieve service logs: %s", err)
	}
	defer io.Close()

	data, err := ioutil.ReadAll(io)
	if err != nil {
		return "", fmt.Errorf("failed to read service logs: %s", err)
	}

	return string(data), nil
}
