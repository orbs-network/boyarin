package adapter

import (
	"context"
	"fmt"
	"path"
)

func (d *dockerAPI) PrepareReverseProxy(ctx context.Context, config string) (Runner, error) {
	image := "nginx:latest"
	if err := d.PullImage(ctx, image); err != nil {
		return nil, fmt.Errorf("could not pull latest nginx image: %s", err)
	}

	nginxConfigDir := path.Join(d.root, "reverse-proxy")
	if err := storeNginxConfiguration(nginxConfigDir, config); err != nil {
		return nil, err
	}

	configMap := make(map[string]interface{})
	configMap["Image"] = image

	exposedPorts := make(map[string]interface{})
	exposedPorts["80/tcp"] = struct{}{}

	configMap["ExposedPorts"] = exposedPorts

	portBindings := make(map[string][]dockerPortBinding)
	portBindings["80/tcp"] = []dockerPortBinding{{"0.0.0.0", "80"}}

	hostConfigMap := make(map[string]interface{})
	hostConfigMap["Binds"] = []string{
		nginxConfigDir + ":/etc/nginx/conf.d",
	}
	hostConfigMap["PortBindings"] = portBindings

	configMap["HostConfig"] = hostConfigMap

	return &dockerRunner{
		client:        d.client,
		config:        configMap,
		containerName: PROXY_CONTAINER_NAME,
	}, nil
}
