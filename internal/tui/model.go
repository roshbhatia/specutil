package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
)

// view selects which panel is shown; the user toggles with tab.
type view int

const (
	viewKanban view = iota
	viewGraph
)

// Model is the root bubbletea model. It holds already-parsed local state and
// never performs I/O during its lifetime.
type Model struct {
	changes []*ir.Change
	graph   *graph.Graph
	diags   []graph.Diagnostic

	byName   map[string]*ir.Change
	columns  map[Lifecycle][]*ir.Change
	view     view
	selected int // index into the flat, lifecycle-ordered change list
	flat     []*ir.Change

	width, height int
}

// New builds a Model from loaded changes and the cross-change graph. The graph
// may be nil/empty when no manifest exists; the kanban still works.
func New(changes []*ir.Change, g *graph.Graph, diags []graph.Diagnostic) Model {
	m := Model{
		changes: changes,
		graph:   g,
		diags:   diags,
		byName:  make(map[string]*ir.Change, len(changes)),
		columns: make(map[Lifecycle][]*ir.Change),
	}
	for _, c := range changes {
		m.byName[c.Name] = c
		lc := Classify(c)
		m.columns[lc] = append(m.columns[lc], c)
	}
	// Flat selection order follows the column order for predictable navigation.
	for _, lc := range LifecycleOrder {
		m.flat = append(m.flat, m.columns[lc]...)
	}
	return m
}

// Init satisfies tea.Model; there is no startup command.
func (m Model) Init() tea.Cmd { return nil }

// Update handles keyboard, mouse, and resize messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			if m.view == viewKanban {
				m.view = viewGraph
			} else {
				m.view = viewKanban
			}
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < len(m.flat)-1 {
				m.selected++
			}
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			for i, c := range m.flat {
				if zone.Get("card:" + c.Name).InBounds(msg) {
					m.selected = i
				}
			}
		}
	}
	return m, nil
}

var (
	styleColumn = lipgloss.NewStyle().Padding(0, 1)
	styleHeader = lipgloss.NewStyle().Bold(true).Underline(true)
	styleCard   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Margin(0, 0, 1, 0)
	styleActive = styleCard.BorderForeground(lipgloss.Color("205"))
	styleHint   = lipgloss.NewStyle().Faint(true)
	styleWarn   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

// View renders the active panel. The whole frame is wrapped in zone.Scan so
// bubblezone can resolve mouse coordinates back to marked card zones.
func (m Model) View() string {
	if len(m.changes) == 0 {
		return zone.Scan(m.emptyState())
	}
	var body string
	switch m.view {
	case viewGraph:
		body = m.graphView()
	default:
		body = m.kanbanView()
	}
	return zone.Scan(body + "\n" + m.footer())
}

func (m Model) emptyState() string {
	return styleWarn.Render("No OpenSpec changes found.") + "\n\n" +
		styleHint.Render("Create one under openspec/changes/<name>/ with a proposal.md, "+
			"then re-run `specutil tui`.") + "\n"
}

func (m Model) footer() string {
	v := "kanban"
	if m.view == viewGraph {
		v = "graph"
	}
	return styleHint.Render(fmt.Sprintf("[%s]  tab: switch view  ·  ←/→: select  ·  q: quit", v))
}

// kanbanView lays out lifecycle columns side by side, each card a workstream.
func (m Model) kanbanView() string {
	var cols []string
	for _, lc := range LifecycleOrder {
		items := m.columns[lc]
		header := styleHeader.Render(fmt.Sprintf("%s (%d)", lc, len(items)))
		cards := []string{header, ""}
		for _, c := range items {
			cards = append(cards, m.card(c))
		}
		cols = append(cols, styleColumn.Render(lipgloss.JoinVertical(lipgloss.Left, cards...)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// card renders a single workstream with its progress, highlighted if selected.
func (m Model) card(c *ir.Change) string {
	done, total := Progress(c)
	label := fmt.Sprintf("%s\n%d/%d tasks", c.Name, done, total)
	style := styleCard
	if len(m.flat) > 0 && m.selected < len(m.flat) && m.flat[m.selected] == c {
		style = styleActive
	}
	return zone.Mark("card:"+c.Name, style.Render(label))
}

// graphView renders the dependency DAG as depth-ordered columns, with cycle and
// dangling diagnostics surfaced as text rather than drawn edges.
func (m Model) graphView() string {
	if m.graph == nil || len(m.graph.Nodes) == 0 {
		return styleHint.Render("No dependency graph. Add edges in openspec/specutil.yaml " +
			"(or run `specutil graph --suggest`).")
	}
	cols := layers(m.graph)
	var rendered []string
	for depth, nodes := range cols {
		header := styleHeader.Render(fmt.Sprintf("depth %d", depth))
		cells := []string{header, ""}
		for _, n := range nodes {
			cells = append(cells, m.card(m.byName[n.ID]))
		}
		rendered = append(rendered, styleColumn.Render(lipgloss.JoinVertical(lipgloss.Left, cells...)))
	}
	out := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	if edges := m.edgeList(); edges != "" {
		out += "\n\n" + styleHint.Render("edges:\n") + edges
	}
	for _, d := range m.diags {
		out += "\n" + styleWarn.Render(fmt.Sprintf("%s: %s", d.Kind, d.Msg))
	}
	return out
}

// edgeList prints the prerequisite -> dependent relations the columns imply but
// can't draw, so the dependency direction stays legible.
func (m Model) edgeList() string {
	if m.graph == nil {
		return ""
	}
	lines := make([]string, 0, len(m.graph.Edges))
	for _, e := range m.graph.Edges {
		lines = append(lines, fmt.Sprintf("  %s → %s", e.From, e.To))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}
