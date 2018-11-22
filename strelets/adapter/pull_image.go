package adapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func pullImage(ctx context.Context, client *client.Client, imageName string) error {
	out, err := client.ImagePull(ctx, imageName, getPullOptions(imageName))

	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	return nil
}

func getPullOptions(imageName string) types.ImagePullOptions {
	if username, password, err := getAuthForRepository(os.Getenv("HOME"), imageName); err != nil {
		fmt.Println(err)
		return types.ImagePullOptions{}
	} else {
		return types.ImagePullOptions{
			RegistryAuth: encodeAuthConfig(&types.AuthConfig{
				Username:      username,
				Password:      password,
				ServerAddress: getRepoName(imageName),
			}),
		}
	}
}

func encodeAuthConfig(auth *types.AuthConfig) string {
	rawJson, _ := json.Marshal(auth)
	return base64.StdEncoding.EncodeToString(rawJson)
}

func getRepoName(imageName string) string {
	if !strings.Contains(imageName, "/") {
		return "registry.docker.io"
	}

	return strings.Split(imageName, "/")[0]
}

func getAuthForRepository(home string, imageName string) (username string, password string, err error) {
	repo := getRepoName(imageName)
	if raw, err := ioutil.ReadFile(path.Join(home, ".docker", "config.json")); err == nil {
		config := make(map[string]interface{})
		json.Unmarshal(raw, &config)

		if auths, ok := config["auths"].(map[string]interface{}); ok {
			if auth, ok := auths[repo].(map[string]interface{}); ok {
				if auth["auth"] == nil {
					return "", "", fmt.Errorf("docker credentials not found for image %s", imageName)
				}

				if encodedUserAndPassword := auth["auth"].(string); ok {
					rawCredentials, err := base64.StdEncoding.DecodeString(encodedUserAndPassword)
					if err != nil {
						return "", "", fmt.Errorf("malformed credentials for image %s", imageName)
					}

					credentials := strings.Split(string(rawCredentials), ":")
					if len(credentials) != 2 {
						return "", "", fmt.Errorf("malformed credentials for image %s", imageName)
					}

					return credentials[0], credentials[1], nil
				}

			}
		}
	}

	return "", "", fmt.Errorf("docker credentials not found for image %s", imageName)
}
