package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
)

// Run starts the interactive TUI over the given changes and dependency graph.
// It blocks until the user quits. Mouse zones are enabled so kanban cards are
// clickable.
func Run(changes []*ir.Change, g *graph.Graph, diags []graph.Diagnostic) error {
	zone.NewGlobal()
	p := tea.NewProgram(
		New(changes, g, diags),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
