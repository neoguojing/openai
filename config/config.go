package config

import (
	"io/ioutil"
	"sync"

	"github.com/neoguojing/log"

	"gopkg.in/yaml.v2"
)

const (
	EnvFilePath string = "FILE_PATH"
	EnvDBPath   string = "DB_PATH"
	EnvLogPath  string = "LOG_PATH"
)

var (
	config *Config
	once   sync.Once
)

type OpenAIConfig struct {
	ApiKey string `yaml:"api_key"`
	Role   string `yaml:"role"`
	Proxy  string `yaml:"proxy"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type AISpeechConfig struct {
	Token          string `yaml:"token"`
	AppID          string `yaml:"appid"`
	EncodingAESKey string `yaml:"aeskey"`
}

type OfficeAccountConfig struct {
	Token          string `yaml:"token"`
	AppID          string `yaml:"appid"`
	AppSecret      string `yaml:"appsecret"`
	EncodingAESKey string `yaml:"aeskey"`
}

type BaiduConfig struct {
	Key    string `yaml:"key"`
	Secret string `yaml:"secret"`
}

type BardConfig struct {
	Token string `yaml:"token"`
}

type ClaudeConfig struct {
	ApiKey string `yaml:"api_key"`
}
type Server struct {
	Port int `yaml:"port"`
}

type Config struct {
	OpenAI        OpenAIConfig        `yaml:"openai"`
	Server        Server              `yaml:"server"`
	Telegram      TelegramConfig      `yaml:"telegram"`
	AISpeech      AISpeechConfig      `yaml:"aispeech"`
	OfficeAccount OfficeAccountConfig `yaml:"officeaccount"`
	Baidu         BaiduConfig         `yaml:"baidu"`
	Bard          BardConfig          `yaml:"bard"`
	Claude        ClaudeConfig        `yaml:"claude"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("yamlFile.Get err   #%v ", err)
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Errorf("Unmarshal: %v", err)
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
			panic(err.Error())
		}
	})

	return config
}
