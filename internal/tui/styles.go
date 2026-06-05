package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("#00C853")
	colorYellow = lipgloss.Color("#FFD600")
	colorRed    = lipgloss.Color("#FF1744")
	colorGray   = lipgloss.Color("#626262")
	colorBlue   = lipgloss.Color("#448AFF")
	colorWhite  = lipgloss.Color("#FAFAFA")
	colorBg     = lipgloss.Color("#1A1A2E")
	colorBgTab  = lipgloss.Color("#16213E")
	colorAccent = lipgloss.Color("#0F3460")

	styleBase = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorWhite)

	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorAccent).
			Padding(0, 2)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorGray).
				Background(colorBgTab).
				Padding(0, 2)

	styleTabBar = lipgloss.NewStyle().
			Background(colorBgTab).
			PaddingTop(1)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			PaddingBottom(1)

	styleCellOK = lipgloss.NewStyle().Foreground(colorGreen)
	styleCellWarn = lipgloss.NewStyle().Foreground(colorYellow)
	styleCellErr  = lipgloss.NewStyle().Foreground(colorRed)
	styleCellGray = lipgloss.NewStyle().Foreground(colorGray)
	styleCellBlue = lipgloss.NewStyle().Foreground(colorBlue)

	styleSelectedRow = lipgloss.NewStyle().
				Background(colorAccent).
				Foreground(colorWhite)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorGray).
			PaddingTop(1)

	styleStatusBar = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(colorWhite).
			Padding(0, 1)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent)
)
