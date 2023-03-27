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
	Value    float64      `yaml:"value"`
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

func ParseArgs(args []string) (*Metrics, *Config, error) {
	flagset := flag.NewFlagSet("config", flag.ContinueOnError)
	listener := "0.0.0.0:2199"
	flagset.StringVar(&listener, "listener", listener, "listener address and port")
	configFile := "config.yml"
	flagset.StringVar(&configFile, "config", configFile, "configuration file")
	err := flagset.Parse(args)
	if err != nil {
		fmt.Println("wrong parameters")
		return nil, nil, err
	}
	metrics, err := ParseMetrics(configFile)
	return metrics, &Config{Listener: listener}, err
}
