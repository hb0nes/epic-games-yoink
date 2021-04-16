package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var config Config

// Config exposes config.yaml
type Config struct {
	HCaptchaURLs      []string `yaml:"hCaptchaURLs"`
	Users             []User   `yaml:"users"`
	Username          string   `yaml:"username"`
	Password          string   `yaml:"password"`
	OTPSecret         string   `yaml:"OTPSecret"`
	TelegramID        string   `yaml:"telegramID"`
	ImgurClientID     string   `yaml:"imgurClientID"`
	ImgurSecret       string   `yaml:"imgurSecret"`
	ImgurRefreshToken string   `yaml:"imgurRefreshToken"`
}

type User struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	OTPSecret  string `yaml:"OTPSecret"`
	TelegramID string `yaml:"telegramID"`
}

func handleErrorFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
