package config

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	config *Config
	once   sync.Once
)

type OpenAIConfig struct {
	ApiKey string `yaml:"api_key"`
	Role   string `yaml:"role"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type Server struct {
	Port int `yaml:"port"`
}

type Config struct {
	OpenAI   OpenAIConfig   `yaml:"openai"`
	Server   Server         `yaml:"server"`
	Telegram TelegramConfig `yaml:"telegram"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return nil, err
	}

	return config, nil
}

func GetConfig() *Config {
	once.Do(func() {
		configPath := "config.yaml"
		var err error
		config, err = LoadConfig(configPath)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	return config
}
