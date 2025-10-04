package history

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/gaetanlhf/SaSHa/internal/config"
)

type Entry struct {
	Fingerprint string    `json:"fingerprint"`
	Timestamp   time.Time `json:"timestamp"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

func makeFingerprint(name, host string, path []string) string {
	key := name + ":" + host + ":" + strings.Join(path, "/")
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])[:16]
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
		Fingerprint: makeFingerprint(server.Name, server.Host, path),
		Timestamp:   time.Now(),
	}

	historyData.Entries = append(historyData.Entries, newEntry)

	if len(historyData.Entries) > historySize {
		historyData.Entries = historyData.Entries[len(historyData.Entries)-historySize:]
	}

	return Save(historyPath, historyData)
}

func FilterByExistingServers(historyData Data, cfg *config.Config) Data {
	var filteredEntries []Entry

	serverIndex := buildServerIndex(cfg)

	for _, entry := range historyData.Entries {
		if _, exists := serverIndex[entry.Fingerprint]; exists {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return Data{Entries: filteredEntries}
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
