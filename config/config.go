package config

type Config struct {
	Docker        DockerConfig
	DockerMachine DockerMachineConfig
}

type DockerConfig struct {
	Endpoint string
}

type DockerMachineConfig struct {
	ReadFromEnv bool
}
