package config

type Config struct {
	Docker DockerConfig
}

type DockerConfig struct {
	Endpoint string
}
