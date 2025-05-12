package main

import (
	"encoding/json"
	"os"
)

type FavoriteEntry struct {
	Server Server   `json:"server"`
	Path   []string `json:"path"`
}

type FavoritesData struct {
	Entries []FavoriteEntry `json:"entries"`
}

func loadFavorites() (FavoritesData, error) {
	var favoritesData FavoritesData

	favoritesPath, err := getFavoritesFilePath()
	if err != nil {
		return favoritesData, err
	}

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

func saveFavorites(favoritesData FavoritesData) error {
	favoritesPath, err := getFavoritesFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(favoritesData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(favoritesPath, data, 0644)
}

func clearFavorites() error {
	favoritesPath, err := getFavoritesFilePath()
	if err != nil {
		return err
	}

	emptyFavorites := FavoritesData{Entries: []FavoriteEntry{}}
	data, err := json.MarshalIndent(emptyFavorites, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(favoritesPath, data, 0644)
}

func addToFavorites(server *Server, path []string, config *Config) error {
	if !config.FavoritesEnabled {
		return nil
	}

	favoritesData, err := loadFavorites()
	if err != nil {
		return err
	}

	for i, entry := range favoritesData.Entries {
		if entry.Server.Name == server.Name && entry.Server.Host == server.Host {
			favoritesData.Entries = append(favoritesData.Entries[:i], favoritesData.Entries[i+1:]...)
			return saveFavorites(favoritesData)
		}
	}

	newEntry := FavoriteEntry{
		Server: *server,
		Path:   append([]string{}, path...),
	}

	favoritesData.Entries = append(favoritesData.Entries, newEntry)
	return saveFavorites(favoritesData)
}

func filterFavoritesByExistingServers(favoritesData FavoritesData, config *Config) FavoritesData {
	var filteredEntries []FavoriteEntry

	allServers := getAllServersFromConfig(config)

	for _, entry := range favoritesData.Entries {
		if serverExistsInMap(entry.Server.Name, entry.Server.Host, allServers) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return FavoritesData{Entries: filteredEntries}
}

func isServerFavorited(server *Server, favoritesData FavoritesData) bool {
	for _, entry := range favoritesData.Entries {
		if entry.Server.Name == server.Name && entry.Server.Host == server.Host {
			return true
		}
	}
	return false
}
