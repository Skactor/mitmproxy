package config

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
)

type ExporterConfig struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}
type ServerConfig struct {
	CertPath string `yaml:"cert"`
	KeyPath  string `yaml:"key"`
	Address  string `yaml:"address"`
}

type Config struct {
	Exporter ExporterConfig `yaml:"exporter"`
	Server   ServerConfig   `yaml:"server"`
}

func Parse(configPath string) (cfg *Config, err error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &cfg)
	return
}
