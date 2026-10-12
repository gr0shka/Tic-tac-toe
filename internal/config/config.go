package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrConfigNotFound = errors.New("not found in config file")
)

type Config struct {
	App      App
	Postgres Postgres
}

type App struct {
	Port string
	Host string
}

type Postgres struct {
	User      string
	Password  string
	Host      string
	Port      string
	Database  string
	Migration string
}

func (p Postgres) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Database)
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	app := App{
		Port: os.Getenv("APP_PORT"),
		Host: os.Getenv("APP_HOST"),
	}

	postgres := Postgres{
		User:      os.Getenv("POSTGRES_USER"),
		Password:  os.Getenv("POSTGRES_PASSWORD"),
		Host:      os.Getenv("POSTGRES_HOST"),
		Port:      os.Getenv("POSTGRES_PORT"),
		Database:  os.Getenv("POSTGRES_DATABASE"),
		Migration: os.Getenv("MIGRATION"),
	}

	switch {
	case app.Port == "":
		slog.Error("No APP_PORT found")
		return nil, fmt.Errorf("APP_PORT: %w", ErrConfigNotFound)
	case app.Host == "":
		slog.Error("No APP_HOST found")
		return nil, fmt.Errorf("APP_HOST: %w", ErrConfigNotFound)

	case postgres.User == "":
		slog.Error("No POSTGRES_USER found")
		return nil, fmt.Errorf("POSTGRES_USER: %w", ErrConfigNotFound)
	case postgres.Password == "":
		slog.Error("No POSTGRES_PASSWORD found")
		return nil, fmt.Errorf("POSTGRES_PASSWORD: %w", ErrConfigNotFound)
	case postgres.Host == "":
		slog.Error("No POSTGRES_HOST found")
		return nil, fmt.Errorf("POSTGRES_HOST: %w", ErrConfigNotFound)
	case postgres.Port == "":
		slog.Error("No POSTGRES_PORT found")
		return nil, fmt.Errorf("POSTGRES_PORT: %w", ErrConfigNotFound)
	case postgres.Database == "":
		slog.Error("No POSTGRES_DATABASE found")
		return nil, fmt.Errorf("POSTGRES_DATABASE: %w", ErrConfigNotFound)
	case postgres.Migration == "":
		log.Print("No MIGRATION found")
		return nil, fmt.Errorf("MIGRATION: %w", ErrConfigNotFound)
	}

	slog.Info("Config successfully loaded")

	return &Config{
		Postgres: postgres,
		App:      app,
	}, nil
}
