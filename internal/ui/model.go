package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/favorites"
	"github.com/gaetanlhf/SaSHa/internal/history"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

func InitialModel(cfg config.Config, historyPath string, favoritesPath string) Model {
	historyEnabled := true
	if cfg.Features.HistorySize == 0 {
		historyEnabled = false
		history.Clear(historyPath)
	} else if cfg.Features.HistorySize < 0 {
		cfg.Features.HistorySize = 20
	}

	currentColor := "#FFFFFF"
	initStyles(currentColor)

	delegate := NewColoredDelegate()
	delegate.DefaultDelegate.Styles.SelectedTitle = selectedItemStyle
	delegate.DefaultDelegate.Styles.SelectedDesc = selectedItemStyle
	delegate.CurrentColor = currentColor
	delegate.InHistoryView = false
	delegate.InFavoritesView = false

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = titleStyle
	l.Styles.HelpStyle = helpStyle

	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.TextStyle = filterTextStyle
	l.FilterInput.Cursor.Style = filterCursorStyle

	startInGroup := cfg.Inventory != nil && len(cfg.Inventory.Groups) == 1 && len(cfg.Inventory.Hosts) == 0
	var initialPath []string
	var items []list.Item
	inErrorView := len(cfg.ImportErrors) > 0

	if inErrorView {
		items = buildErrorItems(cfg.ImportErrors)
		errorDelegate := newErrorDelegate()
		l.SetDelegate(errorDelegate)
	} else if startInGroup {
		singleGroup := cfg.Inventory.Groups[0]
		initialPath = []string{singleGroup.Name}

		currentColor = utils.ResolveEffectiveColor(&cfg, initialPath)
		initStyles(currentColor)
		delegate.CurrentColor = currentColor

		items = buildGroupItems(singleGroup.Groups, initialPath, &cfg)
		for _, server := range singleGroup.Hosts {
			desc := server.Host
			if server.User != nil && *server.User != "" {
				desc = fmt.Sprintf("%s@%s", *server.User, server.Host)
			}
			if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
				desc = fmt.Sprintf("%s:%d", desc, *server.Port)
			}

			serverColor := utils.ResolveEffectiveColor(&cfg, initialPath)
			if server.Color != nil && *server.Color != "" {
				serverColor = *server.Color
			}

			items = append(items, Item{
				Title:       fmt.Sprintf("💻 %s", server.Name),
				Description: desc,
				IsGroup:     false,
				Path:        "",
				Color:       serverColor,
				IsMultiline: false,
			})
		}
	} else {
		if cfg.Inventory != nil {
			items = buildGroupItems(cfg.Inventory.Groups, []string{}, &cfg)
			for _, server := range cfg.Inventory.Hosts {
				if server.Group == "" {
					desc := server.Host
					if server.User != nil && *server.User != "" {
						desc = fmt.Sprintf("%s@%s", *server.User, server.Host)
					}
					if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
						desc = fmt.Sprintf("%s:%d", desc, *server.Port)
					}

					serverColor := utils.ResolveEffectiveColor(&cfg, []string{})
					if server.Color != nil && *server.Color != "" {
						serverColor = *server.Color
					}

					items = append(items, Item{
						Title:       fmt.Sprintf("💻 %s", server.Name),
						Description: desc,
						IsGroup:     false,
						Path:        "",
						Color:       serverColor,
						IsMultiline: false,
					})
				}
			}
		}
	}

	l.SetItems(items)

	helpModel := help.New()
	helpModel.Width = 0
	helpModel.ShowAll = false

	var historyData history.Data
	if historyEnabled {
		historyData, _ = history.Load(historyPath)
	}

	var favoritesData favorites.Data
	if cfg.Features.FavoritesEnabled {
		favoritesData, _ = favorites.Load(favoritesPath)
	}

	keys := newKeyMap(historyEnabled, cfg.Features.FavoritesEnabled, inErrorView)

	var breadcrumbColors []string
	if startInGroup && !inErrorView {
		breadcrumbColors = []string{currentColor}
	}

	return Model{
		Config:           cfg,
		List:             l,
		Help:             helpModel,
		Keys:             keys,
		Breadcrumbs:      []string{},
		BreadcrumbColors: breadcrumbColors,
		GroupStack:       []string{},
		CurrentPath:      initialPath,
		CurrentColor:     currentColor,
		SSHCommand:       "",
		Quitting:         false,
		Width:            0,
		Height:           0,
		InHistoryView:    false,
		InFavoritesView:  false,
		InErrorView:      inErrorView,
		HistoryData:      historyData,
		FavoritesData:    favoritesData,
		StartInGroup:     startInGroup,
		SelectionStack:   []int{},
	}
}

func (m Model) GetSSHCommand() string {
	return m.SSHCommand
}

func (m Model) IsQuitting() bool {
	return m.Quitting
}
