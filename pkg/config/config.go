package config

import (
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort int
	DB      DBConfig
	SMTP    SMTPConfig
	JWT     JWTConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type JWTConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	GoogleClientID string `mapstructure:"GOOGLE_CLIENT_ID"`
}

func LoadConfig() (*Config, error) {
	var c Config

	appPortStr := os.Getenv("APP_PORT")
	port, err := strconv.Atoi(appPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}
	c.AppPort = port

	c.DB.Host = os.Getenv("DB_HOST")
	c.DB.Port = os.Getenv("DB_PORT")
	c.DB.User = os.Getenv("DB_USER")
	c.DB.Password = os.Getenv("DB_PASSWORD")
	c.DB.Name = os.Getenv("DB_NAME")
	c.DB.SSLMode = os.Getenv("DB_SSLMODE")

	c.SMTP.Host = os.Getenv("SMTP_HOST")
	c.SMTP.Port = os.Getenv("SMTP_PORT")
	c.SMTP.Username = os.Getenv("SMTP_USERNAME")
	c.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	c.SMTP.From = os.Getenv("SMTP_FROM")

	c.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	c.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	c.JWT.AccessTTL = 15 * time.Minute
	c.JWT.RefreshTTL = 24 * time.Hour * 7
	c.JWT.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")

	return &c, nil
}
