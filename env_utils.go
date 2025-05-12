package main

import (
	"os"
	"path/filepath"
)

const (
	ENV_SASHA_HOME = "SASHA_HOME"

	CACHE_DIR = "cache"

	CONFIG_FILE    = "config.yaml"
	HISTORY_FILE   = "history.json"
	FAVORITES_FILE = "favorites.json"
)

func getSashaHome() (string, error) {
	homeDir := os.Getenv(ENV_SASHA_HOME)
	if homeDir == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		homeDir = filepath.Join(userHome, ".sasha")
	}

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return "", err
	}

	return homeDir, nil
}

func getCacheDir() (string, error) {
	sashaHome, err := getSashaHome()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(sashaHome, CACHE_DIR)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return cacheDir, nil
}

func getConfigPath() string {
	sashaHome, err := getSashaHome()
	if err != nil {
		return ""
	}

	return filepath.Join(sashaHome, CONFIG_FILE)
}

func getHistoryFilePath() (string, error) {
	sashaHome, err := getSashaHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(sashaHome, HISTORY_FILE), nil
}

func getFavoritesFilePath() (string, error) {
	sashaHome, err := getSashaHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(sashaHome, FAVORITES_FILE), nil
}
