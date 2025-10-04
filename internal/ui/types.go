package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/favorites"
	"github.com/gaetanlhf/SaSHa/internal/history"
)

type Item struct {
	Title          string
	Description    string
	IsGroup        bool
	Path           string
	Color          string
	IsHistory      bool
	HistoryEntry   *history.Entry
	IsFavorite     bool
	FavoriteEntry  *favorites.Entry
	FavoriteStatus bool
	IsMultiline    bool
	PathEntries    []string
	PathColors     []string
	IsError        bool
}

func (i Item) FilterValue() string { return i.Title }

type Model struct {
	Config           config.Config
	List             list.Model
	Help             help.Model
	Keys             KeyMap
	Breadcrumbs      []string
	BreadcrumbColors []string
	GroupStack       []string
	CurrentPath      []string
	CurrentColor     string
	SSHCommand       string
	Quitting         bool
	Width            int
	Height           int
	InHistoryView    bool
	InFavoritesView  bool
	InErrorView      bool
	HistoryData      history.Data
	FavoritesData    favorites.Data
	StartInGroup     bool
	SelectionStack   []int
}

type ColoredDelegate struct {
	DefaultDelegate list.DefaultDelegate
	CurrentColor    string
	InHistoryView   bool
	InFavoritesView bool
}

type errorDelegate struct {
	showDesc bool
}

type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Quit      key.Binding
	Back      key.Binding
	Help      key.Binding
	History   key.Binding
	Favorites key.Binding
	Favorite  key.Binding
	Filter    key.Binding
}
