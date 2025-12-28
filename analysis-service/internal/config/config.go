package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	StorageRoot string
	JWTSecret   string
}

func getDatabaseURL() (string, error) {
	host, ok := os.LookupEnv("DB_HOST")
	if !ok {
		return "", errors.New("Failed to load DB_HOST variable")
	}

	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		return "", errors.New("Failed to load DB_PORT variable")
	}

	user, ok := os.LookupEnv("DB_USER")
	if !ok {
		return "", errors.New("Failed to load DB_USER variable")
	}

	password, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		return "", errors.New("Failed to load DB_PASSWORD variable")
	}

	name, ok := os.LookupEnv("DB_NAME")
	if !ok {
		return "", errors.New("Failed to load DB_NAME variable")
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, name), nil
}

func (c *Config) Load() error {
	var ok bool

	c.Port, ok = os.LookupEnv("PORT")
	if !ok {
		return errors.New("Failed to load PORT variable")
	}

	var err error
	c.DatabaseURL, err = getDatabaseURL()
	if err != nil {
		return err
	}

	c.StorageRoot, ok = os.LookupEnv("STORAGE_ROOT")
	if !ok {
		return errors.New("Failed to load STORAGE_ROOT variable")
	}

	c.JWTSecret, ok = os.LookupEnv("JWT_SECRET")
	if !ok {
		return errors.New("Failed to load JWT_SECRET variable")
	}

	return nil
}
