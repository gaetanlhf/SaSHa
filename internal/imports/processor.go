package imports

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/utils"
	"gopkg.in/yaml.v3"
)

var cacheManager *CacheManager
var orderTracker *OrderTracker

func init() {
	cacheManager = NewCacheManager()
}

type inheritedSettings struct {
	User      *string
	Port      *int
	Password  *string
	ExtraArgs []string
	SSHBinary *string
	Color     *string
	NoCache   bool
	Auth      *config.AuthConfig
}

func Process(cfg *config.Config, configPath string) error {
	orderTracker = LoadOrderTracker(configPath)

	importDirectives := collectImportDirectives(cfg)

	cfg.ImportErrors = []string{}

	if len(importDirectives) > 0 {
		utils.UpdateSpinnerMessage("Processing imports")
		for _, directive := range importDirectives {
			err := processImport(directive, cfg, configPath)
			if err != nil && !strings.Contains(err.Error(), "[EXPIRED_CACHE]") {
				cfg.ImportErrors = append(cfg.ImportErrors, err.Error())
			}
		}
	}

	utils.UpdateSpinnerMessage("Applying inherited settings")
	config.PropagateInheritedSettings(cfg)

	if len(cfg.ImportErrors) > 0 {
		return fmt.Errorf("import errors occurred")
	}
	return nil
}

func collectImportDirectives(cfg *config.Config) []config.ImportDirective {
	var directives []config.ImportDirective

	if cfg.Inventory == nil {
		return directives
	}

	imports := config.GetImportsFromInventory(cfg.Inventory)

	for i := range imports {
		applyGlobalSettingsToDirective(cfg, &imports[i])
	}

	directives = append(directives, imports...)

	for _, group := range cfg.Inventory.Groups {
		groupImports := collectImportDirectivesFromGroup(group, "")
		directives = append(directives, groupImports...)
	}

	return directives
}

func applyGlobalSettingsToDirective(cfg *config.Config, directive *config.ImportDirective) {
	if cfg.Inventory == nil {
		return
	}
	if directive.User == nil && cfg.Inventory.User != nil {
		directive.User = cfg.Inventory.User
	}
	if directive.Port == nil && cfg.Inventory.Port != nil {
		directive.Port = cfg.Inventory.Port
	}
	if directive.Password == nil && cfg.Inventory.Password != nil {
		directive.Password = cfg.Inventory.Password
	}
	if directive.SSHBinary == nil && cfg.Inventory.SSHBinary != nil {
		directive.SSHBinary = cfg.Inventory.SSHBinary
	}
	if len(directive.ExtraArgs) == 0 && len(cfg.Inventory.ExtraArgs) > 0 {
		directive.ExtraArgs = append([]string{}, cfg.Inventory.ExtraArgs...)
	}
	if directive.Auth == nil && cfg.Inventory.Auth != nil {
		directive.Auth = cfg.Inventory.Auth
	}
	if cfg.Inventory.NoCache {
		directive.NoCache = true
	}
}

func collectImportDirectivesFromGroup(group *config.Group, path string) []config.ImportDirective {
	var directives []config.ImportDirective

	currentPath := path
	if currentPath != "" {
		currentPath = currentPath + "/" + group.Name
	} else {
		currentPath = group.Name
	}

	imports := config.GetImportsFromGroup(group)
	groupAuth := config.GetGroupAuth(group)

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

func getGroupInheritedSettings(cfg *config.Config, path string) inheritedSettings {
	if cfg.Inventory == nil {
		return inheritedSettings{}
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return inheritedSettings{}
	}

	var settings inheritedSettings
	var currentGroup *config.Group

	for _, group := range cfg.Inventory.Groups {
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

	settings.Auth = config.GetGroupAuth(currentGroup)

	imports := config.GetImportsFromGroup(currentGroup)
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
				if len(currentGroup.ExtraArgs) > 0 {
					settings.ExtraArgs = append([]string{}, currentGroup.ExtraArgs...)
				}
				if currentGroup.NoCache {
					settings.NoCache = true
				}

				groupAuth := config.GetGroupAuth(currentGroup)
				if groupAuth != nil {
					settings.Auth = groupAuth
				}

				imports := config.GetImportsFromGroup(currentGroup)
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

func applyDirectiveSettingsWithInheritance(importData *config.ImportData, directive config.ImportDirective, inherited inheritedSettings) {
	effectiveUser := inherited.User
	effectivePort := inherited.Port
	effectivePassword := inherited.Password
	effectiveSSHBinary := inherited.SSHBinary
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
		if effectiveNoCache {
			group.NoCache = true
		}
		if len(group.ExtraArgs) == 0 && len(effectiveExtraArgs) > 0 {
			group.ExtraArgs = append([]string{}, effectiveExtraArgs...)
		}

		imports := config.GetImportsFromGroup(group)
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
		if len(host.ExtraArgs) == 0 && len(effectiveExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, effectiveExtraArgs...)
		}
	}
}

func applyAuthToSubgroups(group *config.Group, auth *config.AuthConfig) {
	if auth == nil {
		return
	}

	for _, subgroup := range group.Groups {
		imports := config.GetImportsFromGroup(subgroup)
		for i := range imports {
			if imports[i].Auth == nil {
				imports[i].Auth = auth
			}
		}
		applyAuthToSubgroups(subgroup, auth)
	}
}

func applyDirectiveSettings(importData *config.ImportData, directive config.ImportDirective) {
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
		if directive.NoCache {
			group.NoCache = true
		}
		if len(group.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
			group.ExtraArgs = append([]string{}, directive.ExtraArgs...)
		}

		imports := config.GetImportsFromGroup(group)
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
		if len(host.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, directive.ExtraArgs...)
		}
	}
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func readRemoteFile(urlStr string, schedule string, noCache bool, auth *config.AuthConfig) (config.ImportData, bool, error) {
	var emptyData config.ImportData

	if !noCache && schedule != "" {
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
		if !noCache && schedule != "" {
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
		if !noCache && schedule != "" {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, formatHTTPError(urlStr, err)
			}
		}
		return emptyData, false, formatHTTPError(urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !noCache && schedule != "" {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, fmt.Errorf("HTTP error: %s for %s", resp.Status, urlStr)
			}
		}
		return emptyData, false, fmt.Errorf("HTTP error: %s for %s", resp.Status, urlStr)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if !noCache && schedule != "" {
			cachedData, _, cacheErr := cacheManager.GetCachedData(urlStr, schedule, auth)
			if cacheErr == nil && cachedData != nil {
				return *cachedData, true, fmt.Errorf("failed to read response body from %s: %w", urlStr, err)
			}
		}
		return emptyData, false, fmt.Errorf("failed to read response body from %s: %w", urlStr, err)
	}

	yamlContent := string(data)
	prefix := "importable:"
	if !strings.HasPrefix(yamlContent, prefix) {
		return emptyData, false, fmt.Errorf("invalid import file %s: must start with 'importable:'", urlStr)
	}

	yamlContent = strings.TrimPrefix(yamlContent, prefix)
	yamlContent = strings.TrimLeft(yamlContent, " \t\n\r")
	data = []byte(yamlContent)

	var importData config.ImportData
	if err := yaml.Unmarshal(data, &importData); err != nil {
		if !noCache && schedule != "" {
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

func processImport(directive config.ImportDirective, cfg *config.Config, basePath string) error {
	filePath := directive.File
	resolvedPath := resolveImportPath(filePath, basePath)

	var importData config.ImportData
	var err error
	var usingExpiredCache bool

	if directive.Auth == nil && directive.Path != "" {
		targetGroup := utils.FindGroupByPath(cfg, directive.Path)
		if targetGroup != nil {
			groupAuth := config.GetGroupAuth(targetGroup)
			if groupAuth != nil {
				directive.Auth = groupAuth
			}
		}
	}

	if !directive.NoCache && directive.Path != "" {
		targetGroup := utils.FindGroupByPath(cfg, directive.Path)
		if targetGroup != nil && targetGroup.NoCache {
			directive.NoCache = true
		}
	}

	if isURL(resolvedPath) {
		if !cfg.Features.AllowWebImports {
			errMsg := fmt.Sprintf("Web imports disabled - cannot import %s", filePath)
			return fmt.Errorf(errMsg)
		}
		importData, usingExpiredCache, err = readRemoteFile(resolvedPath, cfg.Features.CacheSchedule, directive.NoCache, directive.Auth)
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

		yamlContent := string(data)
		prefix := "importable:"
		if !strings.HasPrefix(yamlContent, prefix) {
			errMsg := fmt.Sprintf("Invalid import file %s: must start with 'importable:'", filePath)
			return fmt.Errorf(errMsg)
		}

		yamlContent = strings.TrimPrefix(yamlContent, prefix)
		yamlContent = strings.TrimLeft(yamlContent, " \t\n\r")
		data = []byte(yamlContent)

		if yamlErr := yaml.Unmarshal(data, &importData); yamlErr != nil {
			errMsg := fmt.Sprintf("Failed to parse import file %s: %v", filePath, yamlErr)
			return fmt.Errorf(errMsg)
		}
	}

	if usingExpiredCache && err != nil {
		errorMsg := fmt.Sprintf("%v [EXPIRED_CACHE]", err.Error())
		cfg.ImportErrors = append(cfg.ImportErrors, errorMsg)
	}

	for _, group := range importData.Groups {
		nestedImports := config.GetImportsFromGroup(group)
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
			if len(nestedDirective.ExtraArgs) == 0 && len(directive.ExtraArgs) > 0 {
				nestedDirective.ExtraArgs = append([]string{}, directive.ExtraArgs...)
			}
			if directive.NoCache {
				nestedDirective.NoCache = true
			}
			if nestedDirective.Auth == nil && directive.Auth != nil {
				nestedDirective.Auth = directive.Auth
			}

			processImport(nestedDirective, cfg, resolvedPath)
		}
	}

	if directive.Path == "" {
		applyDirectiveSettings(&importData, directive)
		cfg.Inventory.Groups = mergeWithYAMLOrder(cfg.Inventory.Groups, importData.Groups, "", orderTracker)
		cfg.Inventory.Hosts = mergeWithYAMLOrder(cfg.Inventory.Hosts, importData.Hosts, "", orderTracker)
	} else {
		targetGroup := utils.FindGroupByPath(cfg, directive.Path)
		if targetGroup == nil {
			errMsg := fmt.Sprintf("Import target path '%s' not found for %s", directive.Path, filePath)
			return fmt.Errorf(errMsg)
		}

		ancestorSettings := getGroupInheritedSettings(cfg, directive.Path)
		applyDirectiveSettingsWithInheritance(&importData, directive, ancestorSettings)

		targetGroup.Groups = mergeWithYAMLOrder(targetGroup.Groups, importData.Groups, directive.Path, orderTracker)
		targetGroup.Hosts = mergeWithYAMLOrder(targetGroup.Hosts, importData.Hosts, directive.Path, orderTracker)
	}

	return nil
}
