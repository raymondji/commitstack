package config

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	PrimaryColor   lipgloss.Style
	SecondaryColor lipgloss.Style
	TertiaryColor  lipgloss.Style
}

func NewTheme(cfg ThemeConfig) Theme {
	var (
		primaryColor   = "#FFA500"
		secondaryColor = "#00FF00"
		tertiaryColor  = "#9B4DFF"
	)
	if cfg.PrimaryColor != "" {
		primaryColor = cfg.PrimaryColor
	}
	if cfg.SecondaryColor != "" {
		secondaryColor = cfg.SecondaryColor
	}
	if cfg.TertiaryColor != "" {
		secondaryColor = cfg.TertiaryColor
	}

	return Theme{
		PrimaryColor:   lipgloss.NewStyle().Foreground(lipgloss.Color(primaryColor)),
		SecondaryColor: lipgloss.NewStyle().Foreground(lipgloss.Color(secondaryColor)),
		TertiaryColor:  lipgloss.NewStyle().Foreground(lipgloss.Color(tertiaryColor)),
	}
}
