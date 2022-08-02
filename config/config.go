package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AutoStopAfter int64 `yaml:"autoStopAfter"`
	Ffmpeg        struct {
		Command   string `yaml:"command"`
		OutputDir string `yaml:"outputDir"`
	} `yaml:"ffmepg"`
}

func Load(filename string) (*Config, error) {
	c := &Config{}
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %+v", filename, err)
	}

	if err := yaml.Unmarshal([]byte(yamlFile), &c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return c, nil
}
