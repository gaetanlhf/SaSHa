package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
)

type model struct {
	config           Config
	list             list.Model
	help             help.Model
	keys             keyMap
	breadcrumbs      []string
	breadcrumbColors []string
	groupStack       []string
	currentPath      []string
	currentColor     string
	sshCommand       string
	quitting         bool
	width            int
	height           int
	inHistoryView    bool
	inFavoritesView  bool
	inErrorView      bool
	historyData      HistoryData
	favoritesData    FavoritesData
	startInGroup     bool
	selectionStack   []int
}

func initialModel(config Config) model {
	historyEnabled := true
	if config.Features.HistorySize == 0 {
		historyEnabled = false
		clearHistory()
	} else if config.Features.HistorySize < 0 {
		config.Features.HistorySize = 20
	}

	currentColor := "#FFFFFF"
	initStyles(currentColor)

	delegate := NewColoredDelegate()
	delegate.defaultDelegate.Styles.SelectedTitle = selectedItemStyle
	delegate.defaultDelegate.Styles.SelectedDesc = selectedItemStyle
	delegate.currentColor = currentColor
	delegate.inHistoryView = false
	delegate.inFavoritesView = false

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = titleStyle
	l.Styles.HelpStyle = helpStyle

	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.TextStyle = filterTextStyle
	l.FilterInput.Cursor.Style = filterCursorStyle

	startInGroup := len(config.Groups) == 1 && len(config.Hosts) == 0
	var initialPath []string
	var items []list.Item
	inErrorView := len(config.ImportErrors) > 0

	if inErrorView {
		items = buildErrorItems(config.ImportErrors)
		errorDelegate := newErrorDelegate()
		l.SetDelegate(errorDelegate)
	} else if startInGroup {
		singleGroup := config.Groups[0]
		initialPath = []string{singleGroup.Name}

		currentColor = resolveEffectiveColor(&config, initialPath)
		initStyles(currentColor)
		delegate.currentColor = currentColor

		items = buildGroupItems(singleGroup.Groups, initialPath, &config)
		for _, server := range singleGroup.Hosts {
			desc := server.Host
			if server.User != nil && *server.User != "" {
				desc = fmt.Sprintf("%s@%s", *server.User, server.Host)
			}
			if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
				desc = fmt.Sprintf("%s:%d", desc, *server.Port)
			}

			serverColor := resolveEffectiveColor(&config, initialPath)
			if server.Color != nil && *server.Color != "" {
				serverColor = *server.Color
			}

			items = append(items, item{
				title:       fmt.Sprintf("💻 %s", server.Name),
				description: desc,
				isGroup:     false,
				path:        "",
				color:       serverColor,
				isMultiline: false,
			})
		}
	} else {
		items = buildGroupItems(config.Groups, []string{}, &config)
		for _, server := range config.Hosts {
			if server.Group == "" {
				desc := server.Host
				if server.User != nil && *server.User != "" {
					desc = fmt.Sprintf("%s@%s", *server.User, server.Host)
				}
				if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
					desc = fmt.Sprintf("%s:%d", desc, *server.Port)
				}

				serverColor := resolveEffectiveColor(&config, []string{})
				if server.Color != nil && *server.Color != "" {
					serverColor = *server.Color
				}

				items = append(items, item{
					title:       fmt.Sprintf("💻 %s", server.Name),
					description: desc,
					isGroup:     false,
					path:        "",
					color:       serverColor,
					isMultiline: false,
				})
			}
		}
	}

	l.SetItems(items)

	helpModel := help.New()
	helpModel.Width = 0
	helpModel.ShowAll = false

	var historyData HistoryData
	if historyEnabled {
		historyData, _ = loadHistory()
	}

	var favoritesData FavoritesData
	if config.Features.FavoritesEnabled {
		favoritesData, _ = loadFavorites()
	}

	keys := newKeyMap(historyEnabled, config.Features.FavoritesEnabled, inErrorView)

	var breadcrumbColors []string
	if startInGroup && !inErrorView {
		breadcrumbColors = []string{currentColor}
	}

	return model{
		config:           config,
		list:             l,
		help:             helpModel,
		keys:             keys,
		breadcrumbs:      []string{},
		breadcrumbColors: breadcrumbColors,
		groupStack:       []string{},
		currentPath:      initialPath,
		currentColor:     currentColor,
		sshCommand:       "",
		quitting:         false,
		width:            0,
		height:           0,
		inHistoryView:    false,
		inFavoritesView:  false,
		inErrorView:      inErrorView,
		historyData:      historyData,
		favoritesData:    favoritesData,
		startInGroup:     startInGroup,
		selectionStack:   []int{},
	}
}
