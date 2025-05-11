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
	historyData      HistoryData
	favoritesData    FavoritesData
}

func initialModel(config Config) model {
	historyEnabled := true
	if config.HistorySize == 0 {
		historyEnabled = false
		clearHistory()
	} else if config.HistorySize < 0 {
		config.HistorySize = 20
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

	items := buildGroupItems(config.Groups, []string{}, currentColor)
	for _, server := range config.Hosts {
		if server.Group == "" {
			desc := server.Host
			if server.User != "" {
				desc = fmt.Sprintf("%s@%s", server.User, server.Host)
			}
			if server.Port != 0 && server.Port != 22 {
				desc = fmt.Sprintf("%s:%d", desc, server.Port)
			}

			serverColor := server.Color
			if serverColor == "" {
				serverColor = currentColor
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

	l.SetItems(items)

	helpModel := help.New()
	helpModel.Width = 0
	helpModel.ShowAll = false

	var historyData HistoryData
	if historyEnabled {
		historyData, _ = loadHistory()
	}

	var favoritesData FavoritesData
	if config.FavoritesEnabled {
		favoritesData, _ = loadFavorites()
	}

	keys := newKeyMap(historyEnabled, config.FavoritesEnabled)

	return model{
		config:           config,
		list:             l,
		help:             helpModel,
		keys:             keys,
		breadcrumbs:      []string{},
		breadcrumbColors: []string{},
		groupStack:       []string{},
		currentPath:      []string{},
		currentColor:     currentColor,
		sshCommand:       "",
		quitting:         false,
		width:            0,
		height:           0,
		inHistoryView:    false,
		inFavoritesView:  false,
		historyData:      historyData,
		favoritesData:    favoritesData,
	}
}
