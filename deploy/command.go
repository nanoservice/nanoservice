package deploy

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/fsouza/go-dockerclient.v0"

	"github.com/nanoservice/nanoservice/config"
	"github.com/nanoservice/nanoservice/containers"
)

var (
	configPath    string
	configuration *config.Config

	dockerClient *docker.Client
)

const (
	filePermission = 0644
)

func Command(args []string) {
	var err error
	dockerClient, err = containers.NewDockerClient()
	ensureNoError(err, "Unable to configure docker client")
	runApp()
}

func runApp() {
	createServiceNameFile()
	buildApp()
	rmApp()
	startApp()
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
	name := serviceName() + "_1"
	image := serviceName()
	ports := []string{"8080/tcp"}

	ensureNoError(
		containers.Create(dockerClient, image, name, ports),
		"Unable to create container",
	)

	ensureNoError(
		containers.Start(dockerClient, name, ports),
		"Unable to start container "+name,
	)
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

func ensureNoError(err error, message string) {
	if err == nil {
		return
	}

	log.Fatalf("%s: %v", message, err)
}

func currentDir() string {
	dir, err := os.Getwd()
	ensureNoError(err, "Unable to determine current dir")
	return dir
}
