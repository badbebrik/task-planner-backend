package config

import (
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"strconv"
)

type Config struct {
	AppPort int
	DB      DBConfig
	SMTP    SMTPConfig
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

	return &c, nil
}
