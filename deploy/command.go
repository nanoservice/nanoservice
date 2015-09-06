package deploy

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/fsouza/go-dockerclient.v0"

	"github.com/nanoservice/nanoservice/config"
)

var (
	configPath    string
	configuration *config.Config

	dockerClient *docker.Client

	configNotFound = errors.New(
		"Config `.nanoservice.json` not found; try running `nanoservice configure`",
	)
)

const (
	filePermission = 0644
)

func Command(args []string) {
	configPath = findConfig()
	configuration = parseConfig()
	dockerClient = initDockerClient()
	runApp()
}

func runApp() {
	createServiceNameFile()
	buildApp()
	rmApp()
	startApp()
}

func initDockerClient() (client *docker.Client) {
	client, err := docker.NewClient(configuration.Docker.Endpoint)
	ensureNoError(err, "Unable to connect to docker")
	return
}

func buildApp() {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	filepath.Walk(".", filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsDir() {
			return nil
		}

		if strings.HasPrefix(path, ".git") {
			return nil
		}

		log.Printf("transferring %s", path)

		fr, err := os.Open(path)
		ensureNoError(err, "Unable to open file "+path)
		defer fr.Close()

		h, err := tar.FileInfoHeader(info, path)
		ensureNoError(err, "Unable to construct tar info header for "+path)

		err = tw.WriteHeader(h)
		ensureNoError(err, "Unable to write tar headder for "+path)

		_, err = io.Copy(tw, fr)
		ensureNoError(err, "Unable to write file contents to tar "+path)

		return nil
	}))

	tw.Close()

	err := dockerClient.BuildImage(docker.BuildImageOptions{
		Name:         serviceName(),
		InputStream:  buf,
		OutputStream: os.Stdout,
	})

	ensureNoError(err, "Unable to build image")
}

func rmApp() {
	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": []string{
				serviceName(),
			},
		},
	})
	ensureNoError(err, "Unable to verify current status of service")

	if len(containers) == 0 {
		return
	}

	for _, container := range containers {
		err := dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		ensureNoError(err, "Unable to stop already running instance of service")
	}
}

func startApp() {
	hostConfig := &docker.HostConfig{
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8080/tcp": []docker.PortBinding{
				docker.PortBinding{},
			},
		},
	}

	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: serviceName() + "_1",
		Config: &docker.Config{
			Labels: map[string]string{
				serviceName(): "",
			},
			Image: serviceName(),
			ExposedPorts: map[docker.Port]struct{}{
				"8080/tcp": struct{}{},
			},
		},
		HostConfig: hostConfig,
	})
	ensureNoError(err, "Unable to create container")

	err = dockerClient.StartContainer(container.ID, hostConfig)
	ensureNoError(err, "Unable to start container "+container.ID)
}

func rawCommand(name string, cmd ...string) error {
	command := exec.Command(name, cmd...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func rawOutput(name string, cmd ...string) (string, error) {
	var out bytes.Buffer
	command := exec.Command(name, cmd...)
	command.Stdout = &out
	command.Stderr = os.Stderr
	err := command.Run()
	return out.String(), err
}

func createServiceNameFile() {
	ioutil.WriteFile(
		path.Join(currentDir(), ".service_name"),
		[]byte(serviceName()),
		filePermission,
	)
}

func serviceName() string {
	return path.Base(currentDir())
}

func parseConfig() *config.Config {
	data, err := ioutil.ReadFile(configPath)
	ensureNoError(err,
		"Unable to read config file, make sure permissions on `.nanoservice.json` are correct",
	)

	configuration := &config.Config{}
	err = json.Unmarshal(data, configuration)
	ensureNoError(err, "Unable to parse config file")

	return configuration
}

func findConfig() string {
	dir := currentDir()

	for {
		configPath := path.Join(dir, ".nanoservice.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		if dir == "/" {
			break
		}

		dir = path.Join(dir, "..")
	}

	ensureNoError(configNotFound, "Unable to find config")
	return ""
}

func currentDir() string {
	dir, err := os.Getwd()
	ensureNoError(err, "Unable to get current directory")
	return dir
}

func ensureNoError(err error, message string) {
	if err == nil {
		return
	}

	log.Fatalf("%s: %v", message, err)
}
