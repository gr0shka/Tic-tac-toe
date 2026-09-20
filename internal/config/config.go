package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrPasswordNotFound = errors.New("password not found")
	ErrHostNotFound     = errors.New("host not found")
	ErrPortNotFound     = errors.New("port not found")
	ErrDatabaseNotFound = errors.New("database not found")
)

type Config struct {
	Postgres Postgres
}

type Postgres struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func (p Postgres) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Database)
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	postgres := Postgres{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Database: os.Getenv("POSTGRES_DATABASE"),
	}

	switch {
	case postgres.User == "":
		log.Print("No POSTGRES_USER found")
		panic(ErrUserNotFound)
	case postgres.Password == "":
		log.Print("No POSTGRES_PASSWORD found")
		panic(ErrPasswordNotFound)
	case postgres.Host == "":
		log.Print("No POSTGRES_HOST found")
		panic(ErrHostNotFound)
	case postgres.Port == "":
		log.Print("No POSTGRES_PORT found")
		panic(ErrPortNotFound)
	case postgres.Database == "":
		log.Print("No POSTGRES_DATABASE found")
		panic(ErrDatabaseNotFound)
	}

	return &Config{
		Postgres: postgres,
	}, nil
}
