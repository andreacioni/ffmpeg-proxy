package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type FFmpegConfig struct {
	DebugOutput  bool     `yaml:"debugOutput"`
	Command      string   `yaml:"command"`
	Args         []string `yaml:"args"`
	WaitForIndex int64    `yaml:"waitForIndex"`
	IndexFile    string   `yaml:"indexFile"`
}
type Config struct {
	AutoStopAfter int64        `yaml:"autoStopAfter"`
	ServePath     string       `yaml:"servePath"`
	Ffmpeg        FFmpegConfig `yaml:"ffmpeg"`
	Port          int          `yaml:"port"`
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
