package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Title string
	Width int
}

type Row []string

type Table struct {
	Columns  []Column
	Rows     []Row
	Selected int
	Offset   int
	Height   int
}

func NewTable(cols []Column, height int) *Table {
	return &Table{Columns: cols, Height: height}
}

func (t *Table) SetRows(rows []Row) {
	t.Rows = rows
	if t.Selected >= len(rows) {
		t.Selected = max(0, len(rows)-1)
	}
}

func (t *Table) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
		if t.Selected < t.Offset {
			t.Offset = t.Selected
		}
	}
}

func (t *Table) MoveDown() {
	if t.Selected < len(t.Rows)-1 {
		t.Selected++
		if t.Selected >= t.Offset+t.Height {
			t.Offset = t.Selected - t.Height + 1
		}
	}
}

func (t *Table) SelectedRow() Row {
	if t.Selected >= 0 && t.Selected < len(t.Rows) {
		return t.Rows[t.Selected]
	}
	return nil
}

func (t *Table) View() string {
	var b strings.Builder

	// Header
	var headerCells []string
	for _, col := range t.Columns {
		cell := styleHeader.Copy().Width(col.Width).Render(truncate(col.Title, col.Width))
		headerCells = append(headerCells, cell)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Separator
	var sepCells []string
	for _, col := range t.Columns {
		sepCells = append(sepCells, styleCellGray.Copy().Width(col.Width).Render(strings.Repeat("─", col.Width)))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sepCells...))
	b.WriteString("\n")

	// Rows
	end := t.Offset + t.Height
	if end > len(t.Rows) {
		end = len(t.Rows)
	}
	for i := t.Offset; i < end; i++ {
		row := t.Rows[i]
		var cells []string
		for j, col := range t.Columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			cell := lipgloss.NewStyle().Width(col.Width).Render(truncate(val, col.Width))
			if i == t.Selected {
				cell = styleSelectedRow.Copy().Width(col.Width).Render(truncate(val, col.Width))
			}
			cells = append(cells, cell)
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		b.WriteString("\n")
	}

	// Row count
	b.WriteString(fmt.Sprintf("\n%s", styleCellGray.Render(
		fmt.Sprintf("%d/%d", min(t.Selected+1, len(t.Rows)), len(t.Rows)),
	)))

	return b.String()
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
