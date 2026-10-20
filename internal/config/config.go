package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrConfigNotFound = errors.New("not found in config file")
)

type Config struct {
	App      App
	Postgres Postgres
	JWT      JWT
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
	Migration bool
}

type JWT struct {
	PublicKey  string
	PrivateKey string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (p Postgres) ConnectionString() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, p.Port),
		Path:     p.Database,
		RawQuery: "sslmode=disable",
	}

	return u.String()
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	app := App{
		Port: os.Getenv("APP_PORT"),
		Host: os.Getenv("APP_HOST"),
	}

	migration, err := strconv.ParseBool(os.Getenv("MIGRATION"))
	if err != nil {
		slog.Error("parsing MIGRATION environment variable", "err", err)
	}

	postgres := Postgres{
		User:      os.Getenv("POSTGRES_USER"),
		Password:  os.Getenv("POSTGRES_PASSWORD"),
		Host:      os.Getenv("POSTGRES_HOST"),
		Port:      os.Getenv("POSTGRES_PORT"),
		Database:  os.Getenv("POSTGRES_DATABASE"),
		Migration: migration,
	}

	var accessTTL, refreshTTL time.Duration = time.Minute * 15, time.Hour * 24
	aTTL := os.Getenv("JWT_ACCESS_TTL")
	if aTTL == "" {
		slog.Error("JWT_ACCESS_TTL environment variable not set, used default value", "default", aTTL)
	} else {
		accessTTL, err = time.ParseDuration(aTTL)
		if err != nil {
			slog.Error("JWT_ACCESS_TTL environment variable can not be parsed", "err", err)
		}
	}

	rTTL := os.Getenv("JWT_REFRESH_TTL")
	if rTTL == "" {
		slog.Error("JWT_REFRESH_TTL environment variable not set, used default value", "default", rTTL)
	} else {
		refreshTTL, err = time.ParseDuration(rTTL)
		if err != nil {
			slog.Error("JWT_REFRESH_TTL environment variable can not be parsed", "err", err)
		}
	}

	jwt := JWT{
		PublicKey:  os.Getenv("JWT_PUBLIC_KEY"),
		PrivateKey: os.Getenv("JWT_SECRET_KEY"),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
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

	case jwt.PublicKey == "":
		slog.Error("No JWT_PUBLIC_KEY found")
		return nil, fmt.Errorf("JWT_PUBLIC_KEY: %w", ErrConfigNotFound)
	case jwt.PrivateKey == "":
		slog.Error("No JWT_SECRET_KEY found")
		return nil, fmt.Errorf("JWT_SECRET_KEY: %w", ErrConfigNotFound)
	}

	slog.Info("Config successfully loaded")

	return &Config{
		Postgres: postgres,
		App:      app,
		JWT:      jwt,
	}, nil
}
