package favorites

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"

	"github.com/gaetanlhf/SaSHa/internal/config"
)

type Entry struct {
	Fingerprint string `json:"fingerprint"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

func makeFingerprint(name, host string, path []string) string {
	key := name + ":" + host + ":" + strings.Join(path, "/")
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])[:16]
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

	fp := makeFingerprint(server.Name, server.Host, path)

	for i, entry := range favoritesData.Entries {
		if entry.Fingerprint == fp {
			favoritesData.Entries = append(favoritesData.Entries[:i], favoritesData.Entries[i+1:]...)
			return Save(favoritesPath, favoritesData)
		}
	}

	newEntry := Entry{
		Fingerprint: fp,
	}

	favoritesData.Entries = append(favoritesData.Entries, newEntry)
	return Save(favoritesPath, favoritesData)
}

func FilterByExistingServers(favoritesData Data, cfg *config.Config) Data {
	var filteredEntries []Entry

	serverIndex := buildServerIndex(cfg)

	for _, entry := range favoritesData.Entries {
		if _, exists := serverIndex[entry.Fingerprint]; exists {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return Data{Entries: filteredEntries}
}

func IsServerFavorited(server *config.Server, path []string, favoritesData Data) bool {
	fp := makeFingerprint(server.Name, server.Host, path)
	for _, entry := range favoritesData.Entries {
		if entry.Fingerprint == fp {
			return true
		}
	}
	return false
}

type ResolvedEntry struct {
	Server *config.Server
	Path   []string
}

func ResolveEntry(entry Entry, cfg *config.Config) *ResolvedEntry {
	serverIndex := buildServerIndex(cfg)
	return serverIndex[entry.Fingerprint]
}

func buildServerIndex(cfg *config.Config) map[string]*ResolvedEntry {
	index := make(map[string]*ResolvedEntry)

	if cfg.Inventory == nil {
		return index
	}

	for _, server := range cfg.Inventory.Hosts {
		fp := makeFingerprint(server.Name, server.Host, []string{})
		index[fp] = &ResolvedEntry{Server: server, Path: []string{}}
	}

	for _, group := range cfg.Inventory.Groups {
		indexServersFromGroup(group, []string{group.Name}, index)
	}

	return index
}

func indexServersFromGroup(group *config.Group, currentPath []string, index map[string]*ResolvedEntry) {
	for _, server := range group.Hosts {
		fp := makeFingerprint(server.Name, server.Host, currentPath)
		index[fp] = &ResolvedEntry{Server: server, Path: currentPath}
	}

	for _, subgroup := range group.Groups {
		subPath := append([]string{}, currentPath...)
		subPath = append(subPath, subgroup.Name)
		indexServersFromGroup(subgroup, subPath, index)
	}
}
