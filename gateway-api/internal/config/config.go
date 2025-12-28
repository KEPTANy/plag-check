package config

import (
	"errors"
	"os"
)

type Config struct {
	Port                  string
	UserServiceURL        string
	FileStorageServiceURL string
	AnalysisServiceURL    string
}

func (c *Config) Load() error {
	var ok bool

	c.Port, ok = os.LookupEnv("PORT")
	if !ok {
		return errors.New("Failed to load PORT variable")
	}

	c.UserServiceURL, ok = os.LookupEnv("USER_SERVICE_URL")
	if !ok {
		return errors.New("Failed to load USER_SERVICE_URL variable")
	}

	c.FileStorageServiceURL, ok = os.LookupEnv("FILE_STORAGE_SERVICE_URL")
	if !ok {
		return errors.New("Failed to load FILE_STORAGE_SERVICE_URL variable")
	}

	c.AnalysisServiceURL, ok = os.LookupEnv("ANALYSIS_SERVICE_URL")
	if !ok {
		return errors.New("Failed to load ANALYSIS_SERVICE_URL variable")
	}

	return nil
}
