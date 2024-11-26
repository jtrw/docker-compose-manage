package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func LoadConfig(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
