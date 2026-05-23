package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {

	if cfg.DBHost == "" {
		return errors.New("DB_HOST is required")
	}

	if cfg.DBPort == "" {
		return errors.New("DB_PORT is required")
	}

	if cfg.DBUser == "" {
		return errors.New("DB_USER is required")
	}

	if cfg.DBPassword == "" {
		return errors.New("DB_PASSWORD is required")
	}

	if cfg.DBName == "" {
		return errors.New("DB_NAME is required")
	}

	return nil
}
