package configure

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nanoservice/nanoservice/config"
	"log"
	"os"
)

var (
	flags       = flag.NewFlagSet("configure command", flag.ExitOnError)
	commandName = "configure"

	aws    bool
	hosted bool
	docker bool
)

func Command(args []string) {
	parseFlags(args)

	awsIsDefault()

	onlyOnePlatformTypeSpecified()

	unsupported("aws", aws)
	unsupported("hosted", hosted)

	storeDockerConfiguration()
}

func storeDockerConfiguration() {
	configuration := config.Config{
		Docker: config.DockerConfig{
			Endpoint: "unix:///var/run/docker.sock",
		},
	}

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
