package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HCaptchaURLs []string `yaml:"hCaptchaURLs"`
	Username     string   `yaml:"username"`
	Password     string   `yaml:"password"`
}

func readConfig() Config {
	configFile := "./config.yaml"
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Could not open config file.")
	}
	dec := yaml.NewDecoder(f)
	var config Config
	err = dec.Decode(&config)
	if err != nil {
		log.Fatal("Could not open config file.")
	}
	return config
}
