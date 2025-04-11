package configs

import "github.com/spf13/viper"

type Config struct {
	Environment               string `mapstructure:"ENVIRONMENT"`
	BotToken string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	HuggingFaceApiKey string `mapstructure:"HUGGING_FACE_API_KEY"`
}

func GetConfig() (config Config) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var err error

	if err = viper.ReadInConfig(); err != nil {
		panic(err)
	}

	viper.Unmarshal(&config)
	return
}

func IsProd() bool {
	env := GetConfig().Environment

	if env == "production" || env == "prod" {
		return true
	}

	return false
}

func IsDev() bool {
	env := GetConfig().Environment

	if env == "development" || env == "dev" {
		return true
	}

	return false
}

func IsTest() bool {
	env := GetConfig().Environment

	return env == "test"
}