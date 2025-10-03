package history

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gaetanlhf/sasha/internal/config"
)

type Entry struct {
	Server    config.Server `json:"server"`
	Path      []string      `json:"path"`
	Timestamp time.Time     `json:"timestamp"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

func Load(historyPath string) (Data, error) {
	var historyData Data

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

func Save(historyPath string, historyData Data) error {
	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func Clear(historyPath string) error {
	emptyHistory := Data{Entries: []Entry{}}
	data, err := json.MarshalIndent(emptyHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func Add(historyPath string, server *config.Server, path []string, cfg *config.Config) error {
	historySize := cfg.Features.HistorySize
	if historySize == 0 {
		return nil
	}

	if historySize < 0 {
		historySize = 20
	}

	historyData, err := Load(historyPath)
	if err != nil {
		return err
	}

	newEntry := Entry{
		Server:    *server,
		Path:      append([]string{}, path...),
		Timestamp: time.Now(),
	}

	historyData.Entries = append(historyData.Entries, newEntry)

	if len(historyData.Entries) > historySize {
		historyData.Entries = historyData.Entries[len(historyData.Entries)-historySize:]
	}

	return Save(historyPath, historyData)
}

func FilterByExistingServers(historyData Data, cfg *config.Config) Data {
	var filteredEntries []Entry

	allServers := getAllServersFromConfig(cfg)
	validPaths := getAllValidPathsFromConfig(cfg)

	for _, entry := range historyData.Entries {
		if serverExistsInMap(entry.Server.Name, entry.Server.Host, allServers) {
			pathValid := len(entry.Path) == 0
			if len(entry.Path) > 0 {
				for _, validPath := range validPaths {
					if pathsEqual(entry.Path, validPath) {
						pathValid = true
						break
					}
				}
			}
			if pathValid {
				filteredEntries = append(filteredEntries, entry)
			}
		}
	}

	return Data{Entries: filteredEntries}
}

func getAllServersFromConfig(cfg *config.Config) map[string]struct{} {
	serverMap := make(map[string]struct{})

	if cfg.Inventory == nil {
		return serverMap
	}

	for _, server := range cfg.Inventory.Hosts {
		key := server.Name + ":" + server.Host
		serverMap[key] = struct{}{}
	}

	for _, group := range cfg.Inventory.Groups {
		collectServersFromGroup(group, serverMap)
	}

	return serverMap
}

func getAllValidPathsFromConfig(cfg *config.Config) [][]string {
	var paths [][]string

	paths = append(paths, []string{})

	if cfg.Inventory == nil {
		return paths
	}

	for _, group := range cfg.Inventory.Groups {
		collectPathsFromGroup(group, []string{}, &paths)
	}

	return paths
}

func collectPathsFromGroup(group *config.Group, parentPath []string, paths *[][]string) {
	currentPath := append([]string{}, parentPath...)
	currentPath = append(currentPath, group.Name)
	*paths = append(*paths, currentPath)

	for _, subgroup := range group.Groups {
		collectPathsFromGroup(subgroup, currentPath, paths)
	}
}

func pathsEqual(path1, path2 []string) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i := range path1 {
		if path1[i] != path2[i] {
			return false
		}
	}
	return true
}

func collectServersFromGroup(group *config.Group, serverMap map[string]struct{}) {
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
