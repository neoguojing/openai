package config

type OpenAIConfig struct {
	ApiKey string `yaml:"api_key"`
}

type Config struct {
	OpenAI OpenAIConfig `yaml:"openai"`
}
