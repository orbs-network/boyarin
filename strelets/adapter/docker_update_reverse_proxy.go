package adapter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func (d *dockerAPI) UpdateReverseProxy(ctx context.Context, config string) error {
	image := "nginx:latest"
	if err := d.PullImage(ctx, image); err != nil {
		return fmt.Errorf("could not pull latest nginx image: %s", err)
	}

	nginxConfigDir, err := filepath.Abs("_tmp/reverse-proxy")
	if err != nil {
		panic(err)
	}

	os.MkdirAll(nginxConfigDir, 0755)

	if err := ioutil.WriteFile(path.Join(nginxConfigDir, "nginx.conf"), []byte(config), 0644); err != nil {
		return fmt.Errorf("could not save nginx configuration", err)
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

	runner := &dockerRunner{
		client:        d.client,
		config:        configMap,
		containerName: "http-api-reverse-proxy",
	}

	return runner.Run(ctx)
}
