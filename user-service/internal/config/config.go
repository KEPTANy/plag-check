package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	BCryptCost  int
}

func (c *Config) Load() error {
	var ok bool

	c.Port, ok = os.LookupEnv("PORT")
	if !ok {
		return errors.New("Failed to load PORT variable")
	}

	host, ok := os.LookupEnv("DB_HOST")
	if !ok {
		return errors.New("Failed to load DB_HOST variable")
	}

	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		return errors.New("Failed to load DB_PORT variable")
	}

	user, ok := os.LookupEnv("DB_USER")
	if !ok {
		return errors.New("Failed to load DB_USER variable")
	}

	password, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		return errors.New("Failed to load DB_PASSWORD variable")
	}

	name, ok := os.LookupEnv("DB_NAME")
	if !ok {
		return errors.New("Failed to load DB_NAME variable")
	}

	c.DatabaseURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, name)

	c.JWTSecret, ok = os.LookupEnv("JWT_SECRET")
	if !ok {
		return errors.New("Failed to load JWT_SECRET variable")
	}

	var err error
	c.BCryptCost, err = strconv.Atoi(os.Getenv("BCRYPT_COST"))
	if err != nil {
		return errors.New("Failed to load BCRYPT_COST variable")
	}

	return nil
}
