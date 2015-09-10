package containers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/fsouza/go-dockerclient.v0"

	"github.com/nanoservice/nanoservice/config"
)

var (
	configNotFound = errors.New(
		"Config `.nanoservice.json` not found; try running `nanoservice configure`",
	)
)

func Exists(client *docker.Client, ID string) bool {
	_, err := client.InspectContainer(ID)
	return err == nil
}

func Running(client *docker.Client, ID string) (bool, error) {
	container, err := client.InspectContainer(ID)
	if err != nil {
		return false, err
	}

	return container.State.Running, nil
}

func Start(client *docker.Client, ID string, ports []string) error {
	return client.StartContainer(ID, hostConfigFrom(ports, []string{}))
}

func Create(client *docker.Client, image string, name string, label string, ports []string, env []string, links []string) error {
	_, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Labels: map[string]string{
				label: "",
			},
			Image:        image,
			ExposedPorts: exposedPortsFrom(ports),
			Env:          env,
		},
		HostConfig: hostConfigFrom(ports, links),
	})
	return err
}

func NewDockerClient() (*docker.Client, error) {
	configuration, err := fetchConfiguration()
	if err != nil {
		return nil, err
	}

	if configuration.DockerMachine.ReadFromEnv {
		return docker.NewClientFromEnv()
	}

	client, err := docker.NewClient(configuration.Docker.Endpoint)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func hostConfigFrom(ports []string, links []string) *docker.HostConfig {
	return &docker.HostConfig{
		PortBindings: portBindingsFrom(ports),
		Links:        links,
	}
}

func portBindingsFrom(ports []string) map[docker.Port][]docker.PortBinding {
	result := map[docker.Port][]docker.PortBinding{}
	for _, port := range ports {
		result[docker.Port(port)] = []docker.PortBinding{
			docker.PortBinding{},
		}
	}
	return result
}

func exposedPortsFrom(ports []string) map[docker.Port]struct{} {
	result := map[docker.Port]struct{}{}
	for _, port := range ports {
		result[docker.Port(port)] = struct{}{}
	}
	return result
}

func fetchConfiguration() (*config.Config, error) {
	configPath, err := findConfig()
	if err != nil {
		return nil, err
	}

	return parseConfig(configPath)
}

func parseConfig(configPath string) (*config.Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	configuration := &config.Config{}
	if err = json.Unmarshal(data, configuration); err != nil {
		return nil, err
	}

	return configuration, nil
}

func findConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := path.Join(dir, ".nanoservice.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		if dir == "/" {
			break
		}

		dir = path.Join(dir, "..")
	}

	return "", configNotFound
}
