package favorites

import (
	"encoding/json"
	"os"

	"github.com/gaetanlhf/sasha/internal/config"
)

type Entry struct {
	Server config.Server `json:"server"`
	Path   []string      `json:"path"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

func Load(favoritesPath string) (Data, error) {
	var favoritesData Data

	data, err := os.ReadFile(favoritesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return favoritesData, nil
		}
		return favoritesData, err
	}

	if err := json.Unmarshal(data, &favoritesData); err != nil {
		return favoritesData, err
	}

	return favoritesData, nil
}

func Save(favoritesPath string, favoritesData Data) error {
	data, err := json.MarshalIndent(favoritesData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(favoritesPath, data, 0644)
}

func Clear(favoritesPath string) error {
	emptyFavorites := Data{Entries: []Entry{}}
	data, err := json.MarshalIndent(emptyFavorites, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(favoritesPath, data, 0644)
}

func Add(favoritesPath string, server *config.Server, path []string, cfg *config.Config) error {
	if !cfg.Features.FavoritesEnabled {
		return nil
	}

	favoritesData, err := Load(favoritesPath)
	if err != nil {
		return err
	}

	for i, entry := range favoritesData.Entries {
		if entry.Server.Name == server.Name && entry.Server.Host == server.Host {
			favoritesData.Entries = append(favoritesData.Entries[:i], favoritesData.Entries[i+1:]...)
			return Save(favoritesPath, favoritesData)
		}
	}

	newEntry := Entry{
		Server: *server,
		Path:   append([]string{}, path...),
	}

	favoritesData.Entries = append(favoritesData.Entries, newEntry)
	return Save(favoritesPath, favoritesData)
}

func FilterByExistingServers(favoritesData Data, cfg *config.Config) Data {
	var filteredEntries []Entry

	allServers := getAllServersFromConfig(cfg)
	validPaths := getAllValidPathsFromConfig(cfg)

	for _, entry := range favoritesData.Entries {
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

func IsServerFavorited(server *config.Server, favoritesData Data) bool {
	for _, entry := range favoritesData.Entries {
		if entry.Server.Name == server.Name && entry.Server.Host == server.Host {
			return true
		}
	}
	return false
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
