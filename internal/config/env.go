package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Env struct {
	env string `env:"ENV" envDefault:"local"`
	DB  Postgres
}

func Load() (*Env, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		return nil, err
	}

	var env Env
	if err := cleanenv.ReadEnv(&env); err != nil {
		log.Print("Error loading .env file")
		return nil, err
	}

	return &env, nil
}
