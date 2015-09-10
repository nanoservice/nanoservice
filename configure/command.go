package configure

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	dockerlib "gopkg.in/fsouza/go-dockerclient.v0"

	"github.com/nanoservice/nanoservice/config"
	"github.com/nanoservice/nanoservice/containers"
)

var (
	flags       = flag.NewFlagSet("configure command", flag.ExitOnError)
	commandName = "configure"

	aws    bool
	hosted bool
	docker bool

	dockerClient *dockerlib.Client
)

func Command(args []string) {
	parseFlags(args)

	awsIsDefault()

	onlyOnePlatformTypeSpecified()

	unsupported("aws", aws)
	unsupported("hosted", hosted)

	storeDockerConfiguration(newConfiguration())

	var err error
	dockerClient, err = containers.NewDockerClient()
	ensureNoError(err, "Unable to configure docker client")

	runBus()
}

func runBus() {
	if !busExists() {
		createBus()
	}

	if !busRunning() {
		startBus()
	}
}

func startBus() {
	name := busName()
	ports := []string{"2181/tcp", "9092/tcp"}

	ensureNoError(
		containers.Start(dockerClient, name, ports),
		"Unable to start bus cluster",
	)
}

func createBus() {
	name := busName()
	image := "spotify/kafka"
	ports := []string{"2181/tcp", "9092/tcp"}
	env := []string{"ADVERTISED_HOST=bus", "ADVERTISED_PORT=9092"}

	ensureNoError(
		containers.Create(dockerClient, image, name, "bus", ports, env, []string{}),
		"Unable to create bus cluster",
	)
}

func busExists() bool {
	return containers.Exists(dockerClient, busName())
}

func busRunning() bool {
	running, err := containers.Running(dockerClient, busName())
	ensureNoError(err, "Unable to verify if bus is running or not")
	return running
}

func busName() string {
	return "nanoservice_bus_1"
}

func dockerConfiguration() config.Config {
	return config.Config{
		Docker: config.DockerConfig{
			Endpoint: "unix:///var/run/docker.sock",
		},
	}
}

func dockerMachineConfiguration() config.Config {
	return config.Config{
		DockerMachine: config.DockerMachineConfig{
			ReadFromEnv: true,
		},
	}
}

func newConfiguration() config.Config {
	if os.Getenv("DOCKER_MACHINE_NAME") != "" {
		return dockerMachineConfiguration()
	}

	return dockerConfiguration()
}

func storeDockerConfiguration(configuration config.Config) {
	rawConfiguration, err := json.Marshal(configuration)
	if err != nil {
		log.Fatalf("Unable to marshal configuration: %v\n", err)
	}

	file, err := os.Create(".nanoservice.json")
	if err != nil {
		log.Fatalf("Unable to open configuration file: %v\n", err)
	}
	defer file.Close()

	file.Write(rawConfiguration)
}

func unsupported(name string, enabled bool) {
	if !enabled {
		return
	}

	fmt.Printf("--%s is unsupported yet\n", name)
	os.Exit(1)
}

func awsIsDefault() {
	if platformType() != "" {
		return
	}

	aws = true
}

func platformFlagToInt(flag bool) int {
	if !flag {
		return 0
	}
	return 1
}

func onlyOnePlatformTypeSpecified() {
	if platformsSpecified() == 1 {
		return
	}

	fmt.Fprintf(
		os.Stderr,
		"Only one platform should be specified. Got: %d\n\n",
		platformsSpecified(),
	)

	flags.Usage()
}

func platformsSpecified() int {
	return 0 +
		platformFlagToInt(aws) +
		platformFlagToInt(hosted) +
		platformFlagToInt(docker)
}

func platformType() string {
	if aws {
		return "aws"
	}

	if hosted {
		return "hosted"
	}

	if docker {
		return "docker"
	}

	return ""
}

func parseFlags(args []string) {
	flags.BoolVar(&aws, "aws", false, "Configure nanoservice CLI tool for AWS [default] Not implemented yet.")
	flags.BoolVar(&hosted, "hosted", false, "Configure nanoservice CLI tool for hosted setup. Not implemented yet.")
	flags.BoolVar(&docker, "docker", false, "Configure nanoservice CLI tool for local docker setup.")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s %s:\n", os.Args[0], commandName)
		flags.PrintDefaults()
		os.Exit(1)
	}

	flags.Parse(args)
}

func ensureNoError(err error, message string) {
	if err == nil {
		return
	}

	log.Fatalf("%s: %v", message, err)
}
