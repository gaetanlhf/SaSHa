package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/sasha/internal/config"
	"github.com/gaetanlhf/sasha/internal/favorites"
	"github.com/gaetanlhf/sasha/internal/history"
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
