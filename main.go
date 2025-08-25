package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

var (
	version = "dev"
)

func main() {
	clearCacheFlag := flag.Bool("clear-cache", false, "Clear the import cache")
	clearHistoryFlag := flag.Bool("clear-history", false, "Clear connection history")
	clearFavoritesFlag := flag.Bool("clear-favorites", false, "Clear favorites")
	refreshCacheFlag := flag.Bool("refresh-cache", false, "Clear the cache and continue loading")
	printVersion := flag.Bool("version", false, "Print version information")
	help := flag.Bool("help", false, "Show help")
	filterTopLevelGroupsFlag := flag.String("filter-top-level-groups", "", "Only load specified top-level groups (comma-separated)")
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *printVersion {
		fmt.Printf("SaSHa SSH Manager version %s\n", version)
		return
	}

	if *clearCacheFlag {
		err := showSpinner("Clearing cache", func() error {
			return forceCleanCache()
		})
		if err != nil {
			fmt.Printf("Error clearing cache: %v\n", err)
		}
		if !*clearHistoryFlag && !*clearFavoritesFlag {
			return
		}
	}

	if *refreshCacheFlag {
		err := showSpinner("Refreshing cache", func() error {
			return forceCleanCache()
		})
		if err != nil {
			fmt.Printf("Error refreshing cache: %v\n", err)
		}
	}

	if *clearHistoryFlag {
		err := showSpinner("Clearing history", func() error {
			return clearHistory()
		})
		if err != nil {
			fmt.Printf("Error clearing history: %v\n", err)
		}
		if !*clearFavoritesFlag {
			return
		}
	}

	if *clearFavoritesFlag {
		err := showSpinner("Clearing favorites", func() error {
			return clearFavorites()
		})
		if err != nil {
			fmt.Printf("Error clearing favorites: %v\n", err)
		}
		return
	}

	var filterTopLevelGroups []string
	if *filterTopLevelGroupsFlag != "" {
		filterTopLevelGroups = strings.Split(*filterTopLevelGroupsFlag, ",")
		for i := range filterTopLevelGroups {
			filterTopLevelGroups[i] = strings.TrimSpace(filterTopLevelGroups[i])
		}
	}

	configPath := getConfigPath()
	var config Config
	var configErr error

	startSpinner("Loading configuration")
	config, configErr = loadConfig(configPath)
	stopSpinner()

	if configErr != nil && len(config.ImportErrors) == 0 {
		fmt.Printf("Error: %v\n", configErr)
		os.Exit(1)
	}

	if len(filterTopLevelGroups) > 0 {
		filteredConfig, err := filterConfigByGroups(config, filterTopLevelGroups)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		config = filteredConfig
	}

	p := tea.NewProgram(initialModel(config), tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running application:", err)
		os.Exit(1)
	}

	handleApplicationExit(finalModel)
}

func printHelp() {
	fmt.Printf("SaSHa SSH Manager version %s\n\n", version)
	fmt.Println("Usage: sasha [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -clear-cache              Clear the import cache and exit")
	fmt.Println("  -refresh-cache            Clear the cache but continue loading the application")
	fmt.Println("  -clear-history            Clear connection history")
	fmt.Println("  -clear-favorites          Clear favorites")
	fmt.Println("  -filter-top-level-groups  Only load specified top-level groups (comma-separated)")
	fmt.Println("  -version                  Print version information")
	fmt.Println("  -help                     Show this help message")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("  SASHA_HOME                Path to SaSHa home directory (default: ~/.sasha)")
}

func forceCleanCache() error {
	return cacheManager.CleanupCache()
}

func filterConfigByGroups(config Config, filterTopLevelGroups []string) (Config, error) {
	var filteredGroups []*Group
	var foundGroups []string

	for _, groupName := range filterTopLevelGroups {
		found := false
		for _, group := range config.Groups {
			if group.Name == groupName {
				filteredGroups = append(filteredGroups, group)
				foundGroups = append(foundGroups, groupName)
				found = true
				break
			}
		}
		if !found {
			return config, fmt.Errorf("top-level group '%s' not found", groupName)
		}
	}

	filteredConfig := config
	filteredConfig.Groups = filteredGroups

	return filteredConfig, nil
}
