package config

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	PrimaryColor   lipgloss.Style
	SecondaryColor lipgloss.Style
}

func NewTheme(cfg ThemeConfig) Theme {
	var (
		primaryStr   = "#FFA500"
		secondaryStr = "#00FF00"
	)
	if cfg.PrimaryColor != "" {
		primaryStr = cfg.PrimaryColor
	}
	if cfg.SecondaryColor != "" {
		secondaryStr = cfg.SecondaryColor
	}

	return Theme{
		PrimaryColor:   lipgloss.NewStyle().Foreground(lipgloss.Color(primaryStr)),
		SecondaryColor: lipgloss.NewStyle().Foreground(lipgloss.Color(secondaryStr)),
	}
}
