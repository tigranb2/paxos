package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Timeout   int `yaml:"timeout"`
	Acceptors struct {
		Count         int    `yaml:"count"`
		InitIP        string `yaml:"initIP"`
		IncrementIP   bool   `yaml:"incrementIP"`
		InitPort      string `yaml:"initPort"`
		IncrementPort bool   `yaml:"incrementPort"`
	}
	Proposers struct {
		Count         int    `yaml:"count"`
		InitIP        string `yaml:"initIP"`
		IncrementIP   bool   `yaml:"incrementIP"`
		InitPort      string `yaml:"initPort"`
		IncrementPort bool   `yaml:"incrementPort"`
	}
}

//ParseConfig parses config.yaml
func ParseConfig() Config {
	filename, _ := filepath.Abs("./config.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var configData Config

	err = yaml.Unmarshal(yamlFile, &configData)
	if err != nil {
		panic(err)
	}

	return configData
}
