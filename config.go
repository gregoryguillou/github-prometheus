package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listener string
}

type NamedValue struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Bearer struct {
	PersonalAccessToken *string `yaml:"personalAccessToken,omitempty"`
	Endpoint            *string `yaml:"endpoint,omitempty"`
}

type Metric struct {
	Name     string       `yaml:"name"`
	Help     string       `yaml:"help"`
	Bearer   Bearer       `yaml:"bearer"`
	Endpoint string       `yaml:"endpoint"`
	Query    string       `yaml:"query"`
	List     string       `yaml:"list"`
	Labels   []NamedValue `yaml:"labels"`
	Value    interface{}  `yaml:"value"`
}

type Metrics struct {
	Metrics []Metric `yaml:"metrics"`
}

func ParseMetrics(configFile string) (*Metrics, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("could not read file: %s", configFile)
		return nil, err
	}
	metrics := Metrics{}
	err = yaml.Unmarshal([]byte(data), &metrics)
	if err != nil {
		fmt.Printf("could not parse file: %s", configFile)
		return nil, err
	}
	return &metrics, nil
}

func Parse() (*Metrics, *Config, error) {
	listener := "0.0.0.0:2199"
	flag.StringVar(&listener, "listener", listener, "listener address and port")
	configFile := "config.yml"
	flag.StringVar(&configFile, "config", configFile, "configuration file")
	flag.Parse()
	metrics, err := ParseMetrics(configFile)
	return metrics, &Config{Listener: listener}, err
}
