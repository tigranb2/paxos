package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type config struct {
	Proposers struct {
		Count int `yaml:"count"`
		Values []string `yaml:"values"`
	}
	Acceptors struct {
		Count int `yaml:"count"`
		ipRange string `yaml:"ipRange"`
		portRange string `yaml:"portRange"`
	}
}

func parseConfig() config {
	filename, _ := filepath.Abs("./config.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var configData config

	err = yaml.Unmarshal(yamlFile, &configData)
	if err != nil {
		panic(err)
	}

	return configData
}
