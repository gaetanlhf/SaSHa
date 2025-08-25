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
	if directive.User == nil && config.User != nil {
		directive.User = config.User
	}
	if directive.Port == nil && config.Port != nil {
		directive.Port = config.Port
	}
	if directive.Password == nil && config.Password != nil {
		directive.Password = config.Password
	}
	if directive.SSHBinary == nil && config.SSHBinary != nil {
		directive.SSHBinary = config.SSHBinary
	}
	if directive.Color == nil && config.Color != nil {
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
	if group.User == nil && config.User != nil {
		group.User = config.User
	}
	if group.Port == nil && config.Port != nil {
		group.Port = config.Port
	}
	if group.Password == nil && config.Password != nil {
		group.Password = config.Password
	}
	if group.SSHBinary == nil && config.SSHBinary != nil {
		group.SSHBinary = config.SSHBinary
	}
	if group.Color == nil && config.Color != nil {
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
	if server.User == nil && config.User != nil {
		server.User = config.User
	}
	if server.Port == nil && config.Port != nil {
		server.Port = config.Port
	}
	if server.Password == nil && config.Password != nil {
		server.Password = config.Password
	}
	if server.SSHBinary == nil && config.SSHBinary != nil {
		server.SSHBinary = config.SSHBinary
	}
	if server.Color == nil && config.Color != nil {
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

		if imports[i].User == nil && group.User != nil {
			imports[i].User = group.User
		}
		if imports[i].Port == nil && group.Port != nil {
			imports[i].Port = group.Port
		}
		if imports[i].Password == nil && group.Password != nil {
			imports[i].Password = group.Password
		}
		if imports[i].SSHBinary == nil && group.SSHBinary != nil {
			imports[i].SSHBinary = group.SSHBinary
		}
		if imports[i].Color == nil && group.Color != nil {
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
	settings.Password = currentGroup.Password
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

				if currentGroup.User != nil {
					settings.User = currentGroup.User
				}
				if currentGroup.Port != nil {
					settings.Port = currentGroup.Port
				}
				if currentGroup.Password != nil {
					settings.Password = currentGroup.Password
				}
				if currentGroup.SSHBinary != nil {
					settings.SSHBinary = currentGroup.SSHBinary
				}
				if currentGroup.Color != nil {
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
	effectivePassword := inherited.Password
	effectiveSSHBinary := inherited.SSHBinary
	effectiveColor := inherited.Color
	effectiveNoCache := inherited.NoCache || directive.NoCache
	effectiveAuth := inherited.Auth
	effectiveExtraArgs := inherited.ExtraArgs

	if directive.User != nil {
		effectiveUser = directive.User
	}
	if directive.Port != nil {
		effectivePort = directive.Port
	}
	if directive.Password != nil {
		effectivePassword = directive.Password
	}
	if directive.SSHBinary != nil {
		effectiveSSHBinary = directive.SSHBinary
	}
	if directive.Color != nil {
		effectiveColor = directive.Color
	}
	if directive.Auth != nil {
		effectiveAuth = directive.Auth
	}
	if len(directive.ExtraArgs) > 0 {
		effectiveExtraArgs = append([]string{}, directive.ExtraArgs...)
	}

	for _, group := range importData.Groups {
		if group.User == nil && effectiveUser != nil {
			group.User = effectiveUser
		}
		if group.Port == nil && effectivePort != nil {
			group.Port = effectivePort
		}
		if group.Password == nil && effectivePassword != nil {
			group.Password = effectivePassword
		}
		if group.SSHBinary == nil && effectiveSSHBinary != nil {
			group.SSHBinary = effectiveSSHBinary
		}
		if group.Color == nil && effectiveColor != nil {
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
		if host.User == nil && effectiveUser != nil {
			host.User = effectiveUser
		}
		if host.Port == nil && effectivePort != nil {
			host.Port = effectivePort
		}
		if host.Password == nil && effectivePassword != nil {
			host.Password = effectivePassword
		}
		if host.SSHBinary == nil && effectiveSSHBinary != nil {
			host.SSHBinary = effectiveSSHBinary
		}
		if host.Color == nil && effectiveColor != nil {
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
		if group.User == nil && directive.User != nil {
			group.User = directive.User
		}
		if group.Port == nil && directive.Port != nil {
			group.Port = directive.Port
		}
		if group.Password == nil && directive.Password != nil {
			group.Password = directive.Password
		}
		if group.SSHBinary == nil && directive.SSHBinary != nil {
			group.SSHBinary = directive.SSHBinary
		}
		if group.Color == nil && directive.Color != nil {
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
		if host.User == nil && directive.User != nil {
			host.User = directive.User
		}
		if host.Port == nil && directive.Port != nil {
			host.Port = directive.Port
		}
		if host.Password == nil && directive.Password != nil {
			host.Password = directive.Password
		}
		if host.SSHBinary == nil && directive.SSHBinary != nil {
			host.SSHBinary = directive.SSHBinary
		}
		if host.Color == nil && directive.Color != nil {
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
		return fmt.Errorf("⏱️ Connection timeout: %s", urlStr)
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

			if nestedDirective.User == nil && directive.User != nil {
				nestedDirective.User = directive.User
			}
			if nestedDirective.Port == nil && directive.Port != nil {
				nestedDirective.Port = directive.Port
			}
			if nestedDirective.Password == nil && directive.Password != nil {
				nestedDirective.Password = directive.Password
			}
			if nestedDirective.SSHBinary == nil && directive.SSHBinary != nil {
				nestedDirective.SSHBinary = directive.SSHBinary
			}
			if nestedDirective.Color == nil && directive.Color != nil {
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
