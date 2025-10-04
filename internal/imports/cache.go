package imports

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/utils"
	"gopkg.in/yaml.v3"
)

type CacheManager struct {
	cronManager *CronManager
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		cronManager: NewCronManager(),
	}
}

func (c *CacheManager) GetCacheFilePath(url string, auth *config.AuthConfig) (string, error) {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return "", err
	}

	cacheKey := url
	if auth != nil {
		if auth.Username != "" {
			cacheKey = fmt.Sprintf("%s-user:%s", cacheKey, auth.Username)
		}
		if auth.Token != "" {
			tokenHash := md5.Sum([]byte(auth.Token))
			tokenFingerprint := hex.EncodeToString(tokenHash[:])[:8]
			cacheKey = fmt.Sprintf("%s-token:%s", cacheKey, tokenFingerprint)
		}
	}

	hash := md5.Sum([]byte(cacheKey))
	hashStr := hex.EncodeToString(hash[:])

	return filepath.Join(cacheDir, hashStr+".yaml"), nil
}

func (c *CacheManager) SaveToCache(url string, data config.ImportData, auth *config.AuthConfig) error {
	cacheFile, err := c.GetCacheFilePath(url, auth)
	if err != nil {
		return err
	}

	cached := config.CachedImport{
		Metadata: config.CacheMetadata{
			URL:        url,
			LastUpdate: time.Now(),
		},
		Data: data,
	}

	yamlData, err := yaml.Marshal(cached)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, yamlData, 0644)
}

func (c *CacheManager) GetCachedData(url string, schedule string, auth *config.AuthConfig) (*config.ImportData, bool, error) {
	cacheFile, err := c.GetCacheFilePath(url, auth)
	if err != nil {
		return nil, false, err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var cached config.CachedImport
	if err := yaml.Unmarshal(data, &cached); err != nil {
		return nil, false, err
	}

	needsUpdate := c.cronManager.ShouldUpdate(schedule, cached.Metadata.LastUpdate)
	return &cached.Data, needsUpdate, nil
}

func (c *CacheManager) IsCacheExpired(url string, schedule string, auth *config.AuthConfig) bool {
	cacheFile, err := c.GetCacheFilePath(url, auth)
	if err != nil {
		return true
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return true
	}

	var cached config.CachedImport
	if err := yaml.Unmarshal(data, &cached); err != nil {
		return true
	}

	return c.cronManager.ShouldUpdate(schedule, cached.Metadata.LastUpdate)
}

func (c *CacheManager) CleanupCache() error {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			filePath := filepath.Join(cacheDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetCacheManager() *CacheManager {
	return cacheManager
}
