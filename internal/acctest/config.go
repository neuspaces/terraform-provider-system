package acctest

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
)

type Config struct {
	Targets map[string]ConfigTarget `yaml:"targets"`
}

type ConfigTarget struct {
	Os      ConfigTargetOs        `yaml:"os"`
	Configs ConfigTargetConfigMap `yaml:"configs"`
}

type ConfigTargetOs struct {
	Id      string `yaml:"id"`
	Name    string `yaml:"name"`
	Vendor  string `yaml:"vendor"`
	Version string `yaml:"version"`
	Release string `yaml:"release"`
}

type ConfigTargetConfigMap map[string]ConfigTargetConfig

const (
	DefaultTargetConfigId = "default"
)

func (m ConfigTargetConfigMap) Get(id string) (ConfigTargetConfig, error) {
	c, ok := m[id]
	if !ok {
		return ConfigTargetConfig{}, fmt.Errorf("missing target config %q", id)
	}
	return c, nil
}

func (m ConfigTargetConfigMap) MustGet(id string) ConfigTargetConfig {
	c, err := m.Get(id)
	if err != nil {
		panic(err)
	}
	return c
}

func (m ConfigTargetConfigMap) Default() ConfigTargetConfig {
	return m.MustGet(DefaultTargetConfigId)
}

type ConfigTargetConfig struct {
	Ssh ConfigSsh `yaml:"ssh"`
}

type ConfigSsh struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	HostKey    string `yaml:"host_key"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
}

func loadConfig(r io.Reader) (*Config, error) {
	var c Config

	err := yaml.NewDecoder(r).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &c, nil
}
