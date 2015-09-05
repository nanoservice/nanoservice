package deploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/nanoservice/nanoservice/config"
)

var (
	configPath    string
	configuration *config.Config

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
	runApp()
}

func runApp() {
	createServiceNameFile()
	buildApp()
	rmApp()
	startApp()
}

// FIXME: switch to docker client
func buildApp() {
	ensureNoError(
		rawCommand("docker", "build", "-t", serviceName(), "."),
		"Unable to build service",
	)
}

// FIXME: switch to docker client
func rmApp() {
	psOutput, err := rawOutput("docker", "ps", "-a", "-q", "-f", "label="+serviceName())
	ensureNoError(err, "Unable to verify current status of service")

	rawIds := strings.Split(psOutput, "\n")
	ids := []string{}
	for _, id := range rawIds {
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}

	args := append([]string{"rm", "-f"}, ids...)

	ensureNoError(
		rawCommand("docker", args...),
		"Unable to stop currently running service",
	)
}

// FIXME: switch to docker client
func startApp() {
	ensureNoError(
		rawCommand("docker", "run", "-d", "-p", "8080", "--name", serviceName()+"_1", "--label", serviceName(), serviceName()),
		"Unable to start service",
	)
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
