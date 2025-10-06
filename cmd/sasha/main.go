package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/favorites"
	"github.com/gaetanlhf/SaSHa/internal/history"
	"github.com/gaetanlhf/SaSHa/internal/imports"
	"github.com/gaetanlhf/SaSHa/internal/ssh"
	"github.com/gaetanlhf/SaSHa/internal/ui"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

var version = "dev"

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
		fmt.Printf("SaSHa %s\n", version)
		fmt.Printf("https://github.com/gaetanlhf/SaSHa\n")
		return
	}

	historyPath, _ := utils.GetHistoryFilePath()
	favoritesPath, _ := utils.GetFavoritesFilePath()

	if *clearCacheFlag {
		err := utils.ShowSpinner("Clearing cache", func() error {
			return imports.GetCacheManager().CleanupCache()
		})
		if err != nil {
			fmt.Printf("Error clearing cache: %v\n", err)
		}
		if !*clearHistoryFlag && !*clearFavoritesFlag {
			return
		}
	}

	if *refreshCacheFlag {
		err := utils.ShowSpinner("Refreshing cache", func() error {
			return imports.GetCacheManager().CleanupCache()
		})
		if err != nil {
			fmt.Printf("Error refreshing cache: %v\n", err)
		}
	}

	if *clearHistoryFlag {
		err := utils.ShowSpinner("Clearing history", func() error {
			return history.Clear(historyPath)
		})
		if err != nil {
			fmt.Printf("Error clearing history: %v\n", err)
		}
		if !*clearFavoritesFlag {
			return
		}
	}

	if *clearFavoritesFlag {
		err := utils.ShowSpinner("Clearing favorites", func() error {
			return favorites.Clear(favoritesPath)
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

	configPath := utils.GetConfigPath()
	var cfg config.Config
	var configErr error

	utils.StartSpinner("Loading configuration")
	cfg, configErr = config.Load(configPath)

	if configErr == nil {
		configErr = imports.Process(&cfg, configPath)
	}

	utils.StopSpinner()

	if configErr != nil && len(cfg.ImportErrors) == 0 {
		fmt.Printf("Error: %v\n", configErr)
		os.Exit(1)
	}

	if len(filterTopLevelGroups) > 0 {
		filteredConfig, err := config.FilterByGroups(cfg, filterTopLevelGroups)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		cfg = filteredConfig
	}

	ui.SetVersion(version)
	p := tea.NewProgram(ui.InitialModel(cfg, historyPath, favoritesPath), tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running application:", err)
		os.Exit(1)
	}

	handleApplicationExit(finalModel)
}

func printHelp() {
	fmt.Printf("SaSHa %s\n", version)
	fmt.Printf("https://github.com/gaetanlhf/SaSHa \n\n")
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

func handleApplicationExit(finalModel tea.Model) {
	if m, ok := finalModel.(ui.Model); ok && m.IsQuitting() {
		sshCmd := m.GetSSHCommand()
		if sshCmd != "" {
			ssh.Execute(sshCmd)
		}
	}
}
