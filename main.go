package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"path/filepath"
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
		if err := forceCleanCache(); err != nil {
			fmt.Printf("Error clearing cache: %v\n", err)
		} else {
			fmt.Println("Cache cleared successfully")
		}
		if !*clearHistoryFlag && !*clearFavoritesFlag {
			return
		}
	}

	if *refreshCacheFlag {
		if err := forceCleanCache(); err != nil {
			fmt.Printf("Error refreshing cache: %v\n", err)
		} else {
			fmt.Println("Cache refreshed successfully")
		}
	}

	if *clearHistoryFlag {
		if err := clearHistory(); err != nil {
			fmt.Printf("Error clearing history: %v\n", err)
		} else {
			fmt.Println("History cleared successfully")
		}
		if !*clearFavoritesFlag {
			return
		}
	}

	if *clearFavoritesFlag {
		if err := clearFavorites(); err != nil {
			fmt.Printf("Error clearing favorites: %v\n", err)
		} else {
			fmt.Println("Favorites cleared successfully")
		}
		return
	}

	configPath := getConfigPath()
	config, configErr := loadConfig(configPath)

	if configErr != nil && len(config.ImportErrors) > 0 {
		uniqueErrors := dedupErrors(config.ImportErrors)

		if len(uniqueErrors) > 0 {
			fmt.Println("Import Errors:")
			for i, err := range uniqueErrors {
				fmt.Printf("  %d. %s\n", i+1, err)
			}

			fmt.Print("\nSome imports failed. Continue anyway? (y/n) ")
			var answer string
			fmt.Scanln(&answer)

			if answer != "y" && answer != "Y" {
				os.Exit(1)
			}
		}
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
	fmt.Println("  -clear-cache       Clear the import cache and exit")
	fmt.Println("  -refresh-cache     Clear the cache but continue loading the application")
	fmt.Println("  -clear-history     Clear connection history")
	fmt.Println("  -clear-favorites   Clear favorites")
	fmt.Println("  -version           Print version information")
	fmt.Println("  -help              Show this help message")
	fmt.Println("\nEnvironment variables:")
	fmt.Println("  SASHA_HOME         Path to SaSHa home directory (default: ~/.sasha)")
}

func handleApplicationExit(finalModel tea.Model) {
	if m, ok := finalModel.(model); ok && m.quitting && m.sshCommand != "" {
		fmt.Println(m.sshCommand)

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		var shellCommand string

		switch {
		case strings.Contains(shell, "bash"):
			shellCommand = fmt.Sprintf("source ~/.bashrc > /dev/null 2>&1 || true; source ~/.bash_profile > /dev/null 2>&1 || true; %s", m.sshCommand)
		case strings.Contains(shell, "zsh"):
			shellCommand = fmt.Sprintf("source ~/.zshrc > /dev/null 2>&1 || true; %s", m.sshCommand)
		case strings.Contains(shell, "fish"):
			shellCommand = fmt.Sprintf("source ~/.config/fish/config.fish > /dev/null 2>&1 || true; %s", m.sshCommand)
		default:
			shellCommand = m.sshCommand
		}

		cmd := exec.Command(shell, "-i", "-c", shellCommand)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func forceCleanCache() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(cacheDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}
