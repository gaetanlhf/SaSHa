package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func cleanErrorMsg(err string) string {
	err = strings.TrimSpace(err)

	if strings.Contains(err, "Import error for") {
		parts := strings.SplitN(err, "Import error for", 2)
		if len(parts) > 1 {
			err = strings.TrimSpace(parts[1])
		}
	}

	err = strings.Replace(err, "Failed to read import file ", "", 1)

	return err
}

func dedupErrors(errors []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, err := range errors {
		cleanedErr := cleanErrorMsg(err)

		if !seen[cleanedErr] {
			seen[cleanedErr] = true
			result = append(result, err)
		}
	}

	return result
}

func collectImportDirectives(config *Config) []ImportDirective {
	var directives []ImportDirective

	imports := getImportsFromConfig(config)
	directives = append(directives, imports...)

	for _, group := range config.Groups {
		groupImports := collectImportDirectivesFromGroup(group, "")
		directives = append(directives, groupImports...)
	}

	return directives
}

func collectImportDirectivesFromGroup(group *Group, path string) []ImportDirective {
	var directives []ImportDirective

	currentPath := path
	if currentPath != "" {
		currentPath = currentPath + "/" + group.Name
	} else {
		currentPath = group.Name
	}

	imports := getImportsFromGroup(group)
	groupAuth := getGroupAuth(group)

	for i := range imports {
		if imports[i].Path == "" {
			imports[i].Path = currentPath
		}

		if imports[i].User == "" && group.User != "" {
			imports[i].User = group.User
		}

		if imports[i].Port == 0 && group.Port != 0 {
			imports[i].Port = group.Port
		}

		if imports[i].SSHBinary == "" && group.SSHBinary != "" {
			imports[i].SSHBinary = group.SSHBinary
		}

		if imports[i].Color == "" && group.Color != "" {
			imports[i].Color = group.Color
		}

		if len(imports[i].ExtraArgs) == 0 && len(group.ExtraArgs) > 0 {
			imports[i].ExtraArgs = append([]string{}, group.ExtraArgs...)
		}

		if group.NoCache {
			imports[i].NoCache = true
		}

		if imports[i].Auth == nil && groupAuth != nil {
			imports[i].Auth = groupAuth
		}
	}

	directives = append(directives, imports...)

	for _, subgroup := range group.Groups {
		subgroupImports := collectImportDirectivesFromGroup(subgroup, currentPath)

		if groupAuth != nil {
			for i := range subgroupImports {
				if subgroupImports[i].Auth == nil {
					subgroupImports[i].Auth = groupAuth
				}
			}
		}

		directives = append(directives, subgroupImports...)
	}

	return directives
}

func getImportsFromConfig(config *Config) []ImportDirective {
	var importConfig ImportConfig

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil
	}

	err = yaml.Unmarshal(data, &importConfig)
	if err != nil {
		return nil
	}

	return importConfig.Imports
}

func getImportsFromGroup(group *Group) []ImportDirective {
	var importConfig ImportConfig

	data, err := yaml.Marshal(group)
	if err != nil {
		return nil
	}

	err = yaml.Unmarshal(data, &importConfig)
	if err != nil {
		return nil
	}

	return importConfig.Imports
}

func getGroupAuth(group *Group) *AuthConfig {
	data, err := yaml.Marshal(group)
	if err != nil {
		return nil
	}

	type GroupWithAuth struct {
		Auth *AuthConfig `yaml:"auth,omitempty"`
	}

	var groupWithAuth GroupWithAuth
	if err := yaml.Unmarshal(data, &groupWithAuth); err != nil {
		return nil
	}

	return groupWithAuth.Auth
}

func getGroupInheritedSettings(config *Config, path string) inheritedSettings {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return inheritedSettings{}
	}

	var settings inheritedSettings
	var currentGroup *Group

	for _, group := range config.Groups {
		if group.Name == parts[0] {
			currentGroup = group
			break
		}
	}

	if currentGroup == nil {
		return inheritedSettings{}
	}

	settings.User = currentGroup.User
	settings.Port = currentGroup.Port
	settings.SSHBinary = currentGroup.SSHBinary
	settings.Color = currentGroup.Color
	settings.NoCache = currentGroup.NoCache
	if len(currentGroup.ExtraArgs) > 0 {
		settings.ExtraArgs = append([]string{}, currentGroup.ExtraArgs...)
	}

	settings.Auth = getGroupAuth(currentGroup)

	imports := getImportsFromGroup(currentGroup)
	for _, imp := range imports {
		if imp.NoCache {
			settings.NoCache = true
		}
		if settings.Auth == nil && imp.Auth != nil {
			settings.Auth = imp.Auth
		}
	}

	for i := 1; i < len(parts) && currentGroup != nil; i++ {
		found := false
		for _, subgroup := range currentGroup.Groups {
			if subgroup.Name == parts[i] {
				currentGroup = subgroup
				found = true

				if currentGroup.User != "" {
					settings.User = currentGroup.User
				}
				if currentGroup.Port != 0 {
					settings.Port = currentGroup.Port
				}
				if currentGroup.SSHBinary != "" {
					settings.SSHBinary = currentGroup.SSHBinary
				}
				if currentGroup.Color != "" {
					settings.Color = currentGroup.Color
				}
				if len(currentGroup.ExtraArgs) > 0 {
					settings.ExtraArgs = append([]string{}, currentGroup.ExtraArgs...)
				}
				if currentGroup.NoCache {
					settings.NoCache = true
				}

				groupAuth := getGroupAuth(currentGroup)
				if groupAuth != nil {
					settings.Auth = groupAuth
				}

				imports := getImportsFromGroup(currentGroup)
				for _, imp := range imports {
					if imp.NoCache {
						settings.NoCache = true
					}
				}
				break
			}
		}
		if !found {
			break
		}
	}

	return settings
}

func applyDirectiveSettingsWithInheritance(importData *ImportData, directive ImportDirective, inherited inheritedSettings) {
	effectiveUser := inherited.User
	effectivePort := inherited.Port
	effectiveSSHBinary := inherited.SSHBinary
	effectiveColor := inherited.Color
	effectiveNoCache := inherited.NoCache || directive.NoCache
	effectiveAuth := inherited.Auth

	var effectiveExtraArgs []string
	if len(inherited.ExtraArgs) > 0 {
		effectiveExtraArgs = append([]string{}, inherited.ExtraArgs...)
	}

	if directive.User != "" {
		effectiveUser = directive.User
	}
	if directive.Port != 0 {
		effectivePort = directive.Port
	}
	if directive.SSHBinary != "" {
		effectiveSSHBinary = directive.SSHBinary
	}
	if directive.Color != "" {
		effectiveColor = directive.Color
	}
	if directive.Auth != nil {
		effectiveAuth = directive.Auth
	}
	if len(directive.ExtraArgs) > 0 {
		effectiveExtraArgs = append([]string{}, directive.ExtraArgs...)
	}

	for _, group := range importData.Groups {
		if group.User == "" && effectiveUser != "" {
			group.User = effectiveUser
		}
		if group.Port == 0 && effectivePort != 0 {
			group.Port = effectivePort
		}
		if group.SSHBinary == "" && effectiveSSHBinary != "" {
			group.SSHBinary = effectiveSSHBinary
		}
		if group.Color == "" && effectiveColor != "" {
			group.Color = effectiveColor
		}
		if effectiveNoCache {
			group.NoCache = true
		}
		if len(group.ExtraArgs) == 0 && len(effectiveExtraArgs) > 0 {
			group.ExtraArgs = append([]string{}, effectiveExtraArgs...)
		}

		imports := getImportsFromGroup(group)
		for i := range imports {
			if effectiveNoCache {
				imports[i].NoCache = true
			}
			if effectiveAuth != nil && imports[i].Auth == nil {
				imports[i].Auth = effectiveAuth
			}
		}

		applyAuthToSubgroups(group, effectiveAuth)
	}

	for _, host := range importData.Hosts {
		if host.User == "" && effectiveUser != "" {
			host.User = effectiveUser
		}
		if host.Port == 0 && effectivePort != 0 {
			host.Port = effectivePort
		}
		if host.SSHBinary == "" && effectiveSSHBinary != "" {
			host.SSHBinary = effectiveSSHBinary
		}
		if host.Color == "" && effectiveColor != "" {
			host.Color = effectiveColor
		}
		if len(host.ExtraArgs) == 0 && len(effectiveExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, effectiveExtraArgs...)
		}
	}
}

func applyAuthToSubgroups(group *Group, auth *AuthConfig) {
	if auth == nil {
		return
	}

	for _, subgroup := range group.Groups {
		imports := getImportsFromGroup(subgroup)
		for i := range imports {
			if imports[i].Auth == nil {
				imports[i].Auth = auth
			}
		}
		applyAuthToSubgroups(subgroup, auth)
	}
}

func applyDirectiveSettings(importData *ImportData, directive ImportDirective) {
	for _, group := range importData.Groups {
		if group.User == "" && directive.User != "" {
			group.User = directive.User
		}
		if group.Port == 0 && directive.Port != 0 {
			group.Port = directive.Port
		}
		if group.SSHBinary == "" && directive.SSHBinary != "" {
			group.SSHBinary = directive.SSHBinary
		}
		if group.Color == "" && directive.Color != "" {
			group.Color = directive.Color
		}
		if directive.NoCache {
			group.NoCache = true
		}
		if len(group.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
			group.ExtraArgs = append([]string{}, directive.ExtraArgs...)
		}

		imports := getImportsFromGroup(group)
		for i := range imports {
			if directive.NoCache {
				imports[i].NoCache = true
			}
			if directive.Auth != nil && imports[i].Auth == nil {
				imports[i].Auth = directive.Auth
			}
		}
	}

	for _, host := range importData.Hosts {
		if host.User == "" && directive.User != "" {
			host.User = directive.User
		}
		if host.Port == 0 && directive.Port != 0 {
			host.Port = directive.Port
		}
		if host.SSHBinary == "" && directive.SSHBinary != "" {
			host.SSHBinary = directive.SSHBinary
		}
		if host.Color == "" && directive.Color != "" {
			host.Color = directive.Color
		}
		if len(host.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, directive.ExtraArgs...)
		}
	}
}

type ImportCache struct {
	URL       string    `json:"url"`
	Path      string    `json:"path"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func getCacheFilePath(url string, auth *AuthConfig) (string, string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", "", err
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

	metaFile := filepath.Join(cacheDir, hashStr+".meta")
	dataFile := filepath.Join(cacheDir, hashStr+".yaml")

	return metaFile, dataFile, nil
}

func saveToCache(url string, data []byte, auth *AuthConfig) error {
	metaFile, dataFile, err := getCacheFilePath(url, auth)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		return err
	}

	cacheTimeout := 24
	configPath := getConfigPath()
	if configPath != "" {
		if configData, err := os.ReadFile(configPath); err == nil {
			var config Config
			if err := yaml.Unmarshal(configData, &config); err == nil && config.CacheTimeout > 0 {
				cacheTimeout = config.CacheTimeout
			}
		}
	}

	now := time.Now()
	cache := ImportCache{
		URL:       url,
		Path:      dataFile,
		CachedAt:  now,
		ExpiresAt: now.Add(time.Duration(cacheTimeout) * time.Hour),
	}

	metaData, err := yaml.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(metaFile, metaData, 0644)
}

func getCachedFile(url string, auth *AuthConfig) ([]byte, error) {
	metaFile, dataFile, err := getCacheFilePath(url, auth)
	if err != nil {
		return nil, err
	}

	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache ImportCache
	if err := yaml.Unmarshal(metaData, &cache); err != nil {
		return nil, err
	}

	if time.Now().After(cache.ExpiresAt) {
		return nil, nil
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

func cleanupCache() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".meta") {
			metaPath := filepath.Join(cacheDir, file.Name())

			metaData, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}

			var cache ImportCache
			if err := yaml.Unmarshal(metaData, &cache); err != nil {
				continue
			}

			if time.Now().After(cache.ExpiresAt) {
				os.Remove(metaPath)
				os.Remove(cache.Path)
			}
		}
	}

	return nil
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func readRemoteFile(urlStr string, noCache bool, auth *AuthConfig) ([]byte, error) {
	if !noCache {
		cachedData, err := getCachedFile(urlStr, auth)
		if err == nil && cachedData != nil {
			return cachedData, nil
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", urlStr, err)
	}

	req.Header.Set("User-Agent", "SaSHa-SSH-Manager")

	if auth != nil {
		if auth.Username != "" && auth.Password != "" {
			req.SetBasicAuth(auth.Username, auth.Password)
		} else if auth.Token != "" {
			headerName := "Authorization"
			if auth.Header != "" {
				headerName = auth.Header
			}
			req.Header.Set(headerName, "Bearer "+auth.Token)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if !noCache {
			cachedData, _ := getCachedFile(urlStr, auth)
			if cachedData != nil {
				return cachedData, nil
			}
		}
		return nil, formatHTTPError(urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !noCache {
			cachedData, _ := getCachedFile(urlStr, auth)
			if cachedData != nil {
				return cachedData, nil
			}
		}
		return nil, fmt.Errorf("HTTP error: %s for %s", resp.Status, urlStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if !noCache {
			cachedData, _ := getCachedFile(urlStr, auth)
			if cachedData != nil {
				return cachedData, nil
			}
		}
		return nil, fmt.Errorf("failed to read response body from %s: %w", urlStr, err)
	}

	var importData ImportData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		if !noCache {
			cachedData, _ := getCachedFile(urlStr, auth)
			if cachedData != nil {
				return cachedData, nil
			}
		}
		return data, fmt.Errorf("failed to parse imported data from %s: %w", urlStr, err)
	}

	if !noCache {
		saveToCache(urlStr, data, auth)
	}

	return data, nil
}

func formatHTTPError(urlStr string, err error) error {
	switch {
	case strings.Contains(err.Error(), "no such host"):
		return fmt.Errorf("host not found for %s: %w", urlStr, err)
	case strings.Contains(err.Error(), "timeout"):
		return fmt.Errorf("connection timed out for %s: %w", urlStr, err)
	case strings.Contains(err.Error(), "connection refused"):
		return fmt.Errorf("connection refused for %s: %w", urlStr, err)
	case strings.Contains(err.Error(), "no route to host"):
		return fmt.Errorf("no route to host for %s: %w", urlStr, err)
	case strings.Contains(err.Error(), "certificate"):
		return fmt.Errorf("SSL/TLS certificate error for %s: %w", urlStr, err)
	default:
		return fmt.Errorf("error fetching %s: %w", urlStr, err)
	}
}

func resolveImportPath(importPath, basePath string) string {
	if isURL(importPath) {
		return importPath
	}

	if isURL(basePath) {
		baseURL, err := url.Parse(basePath)
		if err != nil {
			return importPath
		}

		if strings.HasPrefix(importPath, "/") {
			baseURL.Path = importPath
			return baseURL.String()
		}

		baseDir := filepath.Dir(baseURL.Path)
		if baseDir == "." {
			baseDir = ""
		}

		newPath := filepath.Join(baseDir, importPath)
		newPath = strings.ReplaceAll(newPath, "\\", "/")

		baseURL.Path = newPath
		return baseURL.String()
	}

	if !filepath.IsAbs(importPath) {
		return filepath.Join(filepath.Dir(basePath), importPath)
	}

	return importPath
}

func ProcessImports(config *Config, configPath string) error {
	_ = cleanupCache()

	importDirectives := collectImportDirectives(config)

	config.ImportErrors = []string{}
	var lastError error

	for _, directive := range importDirectives {
		err := processImport(directive, config, configPath)
		if err != nil {
			config.ImportErrors = append(config.ImportErrors, err.Error())
			lastError = err
		}
	}

	propagateInheritedSettings(config)

	return lastError
}

func processImport(directive ImportDirective, config *Config, basePath string) error {
	filePath := directive.File
	resolvedPath := resolveImportPath(filePath, basePath)

	var data []byte
	var err error

	if directive.Auth == nil && directive.Path != "" {
		targetGroup := findGroupByPath(config, directive.Path)
		if targetGroup != nil {
			groupAuth := getGroupAuth(targetGroup)
			if groupAuth != nil {
				directive.Auth = groupAuth
			}
		}
	}

	if !directive.NoCache && directive.Path != "" {
		targetGroup := findGroupByPath(config, directive.Path)
		if targetGroup != nil && targetGroup.NoCache {
			directive.NoCache = true
		}
	}

	if isURL(resolvedPath) {
		data, err = readRemoteFile(resolvedPath, directive.NoCache, directive.Auth)
	} else {
		data, err = os.ReadFile(resolvedPath)
	}

	if err != nil || data == nil || len(data) == 0 {
		errMsg := fmt.Sprintf("Failed to read import file %s: %v", filePath, err)
		return fmt.Errorf(errMsg)
	}

	var importData ImportData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		errMsg := fmt.Sprintf("Failed to parse import file %s: %v", filePath, err)
		return fmt.Errorf(errMsg)
	}

	for _, group := range importData.Groups {
		nestedImports := getImportsFromGroup(group)
		for _, nestedImport := range nestedImports {
			nestedDirective := nestedImport

			if nestedDirective.Path == "" {
				if directive.Path == "" {
					nestedDirective.Path = group.Name
				} else {
					nestedDirective.Path = filepath.Join(directive.Path, group.Name)
					nestedDirective.Path = strings.ReplaceAll(nestedDirective.Path, "\\", "/")
				}
			}

			if nestedDirective.User == "" && directive.User != "" {
				nestedDirective.User = directive.User
			}
			if nestedDirective.Port == 0 && directive.Port != 0 {
				nestedDirective.Port = directive.Port
			}
			if nestedDirective.SSHBinary == "" && directive.SSHBinary != "" {
				nestedDirective.SSHBinary = directive.SSHBinary
			}
			if nestedDirective.Color == "" && directive.Color != "" {
				nestedDirective.Color = directive.Color
			}
			if len(nestedDirective.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
				nestedDirective.ExtraArgs = append([]string{}, directive.ExtraArgs...)
			}
			if directive.NoCache {
				nestedDirective.NoCache = true
			}

			if nestedDirective.Auth == nil && directive.Auth != nil {
				nestedDirective.Auth = directive.Auth
			}

			processImport(nestedDirective, config, resolvedPath)
		}
	}

	if directive.Path == "" {
		applyDirectiveSettings(&importData, directive)
		config.Groups = append(config.Groups, importData.Groups...)
		config.Hosts = append(config.Hosts, importData.Hosts...)
	} else {
		targetGroup := findGroupByPath(config, directive.Path)
		if targetGroup == nil {
			errMsg := fmt.Sprintf("Import target path '%s' not found for %s", directive.Path, filePath)
			return fmt.Errorf(errMsg)
		}

		ancestorSettings := getGroupInheritedSettings(config, directive.Path)
		applyDirectiveSettingsWithInheritance(&importData, directive, ancestorSettings)

		targetGroup.Groups = append(targetGroup.Groups, importData.Groups...)
		targetGroup.Hosts = append(targetGroup.Hosts, importData.Hosts...)
	}

	return nil
}
