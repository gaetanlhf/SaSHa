package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type HistoryEntry struct {
	Server    Server    `json:"server"`
	Path      []string  `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type HistoryData struct {
	Entries []HistoryEntry `json:"entries"`
}

func getHistoryFilePath() (string, error) {
	configDir := os.Getenv("SASHA_CONFIG_DIR")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".sasha")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "history.json"), nil
}

func loadHistory() (HistoryData, error) {
	var historyData HistoryData

	historyPath, err := getHistoryFilePath()
	if err != nil {
		return historyData, err
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return historyData, nil
		}
		return historyData, err
	}

	if err := json.Unmarshal(data, &historyData); err != nil {
		return historyData, err
	}

	return historyData, nil
}

func saveHistory(historyData HistoryData) error {
	historyPath, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func clearHistory() error {
	historyPath, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	emptyHistory := HistoryData{Entries: []HistoryEntry{}}
	data, err := json.MarshalIndent(emptyHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func addToHistory(server *Server, path []string) error {
	config, err := loadConfig(getConfigPath())
	if err != nil {
		return err
	}

	return addToHistoryWithConfig(server, path, &config)
}

func addToHistoryWithConfig(server *Server, path []string, config *Config) error {
	historySize := config.HistorySize
	if historySize == 0 {
		return nil
	}

	if historySize < 0 {
		historySize = 20
	}

	historyData, err := loadHistory()
	if err != nil {
		return err
	}

	newEntry := HistoryEntry{
		Server:    *server,
		Path:      append([]string{}, path...),
		Timestamp: time.Now(),
	}

	historyData.Entries = append(historyData.Entries, newEntry)

	if len(historyData.Entries) > historySize {
		historyData.Entries = historyData.Entries[len(historyData.Entries)-historySize:]
	}

	return saveHistory(historyData)
}

func filterHistoryByExistingServers(historyData HistoryData, config *Config) HistoryData {
	var filteredEntries []HistoryEntry

	allServers := getAllServersFromConfig(config)

	for _, entry := range historyData.Entries {
		if serverExistsInMap(entry.Server.Name, entry.Server.Host, allServers) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return HistoryData{Entries: filteredEntries}
}

func getAllServersFromConfig(config *Config) map[string]struct{} {
	serverMap := make(map[string]struct{})

	for _, server := range config.Hosts {
		key := server.Name + ":" + server.Host
		serverMap[key] = struct{}{}
	}

	for _, group := range config.Groups {
		collectServersFromGroup(group, serverMap)
	}

	return serverMap
}

func collectServersFromGroup(group *Group, serverMap map[string]struct{}) {
	for _, server := range group.Hosts {
		key := server.Name + ":" + server.Host
		serverMap[key] = struct{}{}
	}

	for _, subgroup := range group.Groups {
		collectServersFromGroup(subgroup, serverMap)
	}
}

func serverExistsInMap(name, host string, serverMap map[string]struct{}) bool {
	key := name + ":" + host
	_, exists := serverMap[key]
	return exists
}

func getConfigPath() string {
	configPath := os.Getenv("SASHA_CONFIG")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configPath = filepath.Join(homeDir, ".sasha", "config.yaml")
	}
	return configPath
}
