package config

import (
	"log"
	"os"
)

type Config struct {
	TelegramToken string
	Debug         bool
}

func Load() *Config {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN не установлен")
	}
	return &Config{
		TelegramToken: token,
		Debug:         true,
	}
}
