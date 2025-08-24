package main

import (
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

var cacheManager *CacheManager

func init() {
	cacheManager = NewCacheManager()
}

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

	for i := range imports {
		applyGlobalSettingsToDirective(config, &imports[i])
	}

	directives = append(directives, imports...)

	for _, group := range config.Groups {
		applyGlobalSettingsToGroup(config, group)

		groupImports := collectImportDirectivesFromGroup(group, "")
		directives = append(directives, groupImports...)
	}

	return directives
}

func applyGlobalSettings(config *Config) {
	for _, server := range config.Hosts {
		applyGlobalSettingsToServer(config, server)
	}

	for _, group := range config.Groups {
		applyGlobalSettingsToGroup(config, group)
	}
}

func applyGlobalSettingsToDirective(config *Config, directive *ImportDirective) {
	if directive.User == "" && config.User != "" {
		directive.User = config.User
	}
	if directive.Port == 0 && config.Port != 0 {
		directive.Port = config.Port
	}
	if directive.SSHBinary == "" && config.SSHBinary != "" {
		directive.SSHBinary = config.SSHBinary
	}
	if directive.Color == "" && config.Color != "" {
		directive.Color = config.Color
	}
	if len(directive.ExtraArgs) == 0 && len(config.ExtraArgs) > 0 {
		directive.ExtraArgs = append([]string{}, config.ExtraArgs...)
	}
	if directive.Auth == nil && config.Auth != nil {
		directive.Auth = config.Auth
	}
	if config.NoCache {
		directive.NoCache = true
	}
}

func applyGlobalSettingsToGroup(config *Config, group *Group) {
	if group.User == "" && config.User != "" {
		group.User = config.User
	}
	if group.Port == 0 && config.Port != 0 {
		group.Port = config.Port
	}
	if group.SSHBinary == "" && config.SSHBinary != "" {
		group.SSHBinary = config.SSHBinary
	}
	if group.Color == "" && config.Color != "" {
		group.Color = config.Color
	}
	if len(group.ExtraArgs) == 0 && len(config.ExtraArgs) > 0 {
		group.ExtraArgs = append([]string{}, config.ExtraArgs...)
	}
	if group.Auth == nil && config.Auth != nil {
		group.Auth = config.Auth
	}
	if config.NoCache {
		group.NoCache = true
	}

	for _, server := range group.Hosts {
		applyGlobalSettingsToServer(config, server)
	}

	for _, subgroup := range group.Groups {
		applyGlobalSettingsToGroup(config, subgroup)
	}
}

func applyGlobalSettingsToServer(config *Config, server *Server) {
	if server.User == "" && config.User != "" {
		server.User = config.User
	}
	if server.Port == 0 && config.Port != 0 {
		server.Port = config.Port
	}
	if server.SSHBinary == "" && config.SSHBinary != "" {
		server.SSHBinary = config.SSHBinary
	}
	if server.Color == "" && config.Color != "" {
		server.Color = config.Color
	}
	if len(server.ExtraArgs) == 0 && len(config.ExtraArgs) > 0 {
		server.ExtraArgs = append([]string{}, config.ExtraArgs...)
	}
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
	effectiveExtraArgs := inherited.ExtraArgs

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

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func readRemoteFile(urlStr string, schedule string, noCache bool, auth *AuthConfig) (ImportData, bool, error) {
	var emptyData ImportData

	if !noCache {
		cachedData, needsUpdate, err := cacheManager.GetCachedData(urlStr, schedule, auth)
		if err == nil && cachedData != nil && !needsUpdate {
			return *cachedData, false, nil
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		if !noCache {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, formatHTTPError(urlStr, err)
			}
		}
		return emptyData, false, formatHTTPError(urlStr, err)
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
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, formatHTTPError(urlStr, err)
			}
		}
		return emptyData, false, formatHTTPError(urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !noCache {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, fmt.Errorf("HTTP error: %s for %s", resp.Status, urlStr)
			}
		}
		return emptyData, false, fmt.Errorf("HTTP error: %s for %s", resp.Status, urlStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if !noCache {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, fmt.Errorf("failed to read response body from %s: %w", urlStr, err)
			}
		}
		return emptyData, false, fmt.Errorf("failed to read response body from %s: %w", urlStr, err)
	}

	var importData ImportData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		if !noCache {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, fmt.Errorf("failed to parse imported data from %s: %w", urlStr, err)
			}
		}
		return emptyData, false, fmt.Errorf("failed to parse imported data from %s: %w", urlStr, err)
	}

	if !noCache {
		cacheManager.SaveToCache(urlStr, importData, auth)
	}

	return importData, false, nil
}

func formatHTTPError(urlStr string, err error) error {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "no such host"):
		return fmt.Errorf("🌐 Host not found: %s", urlStr)
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("ⱱ️ Connection timeout: %s", urlStr)
	case strings.Contains(errStr, "connection refused"):
		return fmt.Errorf("🚫 Connection refused: %s", urlStr)
	case strings.Contains(errStr, "no route to host"):
		return fmt.Errorf("🛣️ No route to host: %s", urlStr)
	case strings.Contains(errStr, "certificate"):
		return fmt.Errorf("🔒 SSL/TLS certificate error: %s", urlStr)
	default:
		return fmt.Errorf("🌐 Network error for %s: %v", urlStr, err)
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
	importDirectives := collectImportDirectives(config)

	config.ImportErrors = []string{}

	if len(importDirectives) > 0 {
		updateSpinnerMessage("Processing imports")
		for _, directive := range importDirectives {
			err := processImport(directive, config, configPath)
			if err != nil && !strings.Contains(err.Error(), "[EXPIRED_CACHE]") {
				config.ImportErrors = append(config.ImportErrors, err.Error())
			}
		}
	}

	updateSpinnerMessage("Applying inherited settings")
	propagateInheritedSettings(config)

	if len(config.ImportErrors) > 0 {
		return fmt.Errorf("import errors occurred")
	}
	return nil
}

func processImport(directive ImportDirective, config *Config, basePath string) error {
	filePath := directive.File
	resolvedPath := resolveImportPath(filePath, basePath)

	var importData ImportData
	var err error
	var usingExpiredCache bool

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
		importData, usingExpiredCache, err = readRemoteFile(resolvedPath, config.CacheSchedule, directive.NoCache, directive.Auth)
		if err != nil && !usingExpiredCache {
			errMsg := fmt.Sprintf("Failed to read import file %s: %v", filePath, err)
			return fmt.Errorf(errMsg)
		}
	} else {
		data, readErr := os.ReadFile(resolvedPath)
		if readErr != nil {
			errMsg := fmt.Sprintf("Failed to read import file %s: %v", filePath, readErr)
			return fmt.Errorf(errMsg)
		}

		if len(data) == 0 {
			errMsg := fmt.Sprintf("Failed to read import file %s: empty file", filePath)
			return fmt.Errorf(errMsg)
		}

		if yamlErr := yaml.Unmarshal(data, &importData); yamlErr != nil {
			errMsg := fmt.Sprintf("Failed to parse import file %s: %v", filePath, yamlErr)
			return fmt.Errorf(errMsg)
		}
	}

	if usingExpiredCache && err != nil {
		errorMsg := fmt.Sprintf("%v [EXPIRED_CACHE]", err.Error())
		config.ImportErrors = append(config.ImportErrors, errorMsg)
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
