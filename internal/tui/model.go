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

// minColWidth is the narrowest a lifecycle/depth column may get before the
// layout degrades to a single scrolling list instead of side-by-side columns.
const minColWidth = 24

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

	detailOpen  bool // master-detail ticket panel visibility
	boardScroll int  // vertical scroll offset for the board
	detailScrol int  // vertical scroll offset for the detail panel

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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Esc closes the ticket panel first; only quits when nothing is open.
			if m.detailOpen {
				m.detailOpen = false
				m.detailScrol = 0
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if len(m.flat) > 0 {
				m.detailOpen = true
				m.detailScrol = 0
			}
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
		case "up", "k":
			m.scroll(-1)
		case "down", "j":
			m.scroll(1)
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			for i, c := range m.flat {
				if zone.Get("card:" + c.Name).InBounds(msg) {
					m.selected = i
					m.detailOpen = true // a click opens the ticket, like a kanban card
					m.detailScrol = 0
				}
			}
		}
	}
	return m, nil
}

// scroll nudges the active pane's offset, clamping at zero. The lower bound is
// enforced lazily at render time against the actual content height.
func (m *Model) scroll(delta int) {
	target := &m.boardScroll
	if m.detailOpen {
		target = &m.detailScrol
	}
	if *target+delta >= 0 {
		*target += delta
	}
}

var (
	styleColumn   = lipgloss.NewStyle().Padding(0, 1)
	styleHeader   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleCard     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Margin(0, 0, 1, 0)
	styleActive   = styleCard.BorderForeground(lipgloss.Color("205"))
	styleNeighbor = styleCard.BorderForeground(lipgloss.Color("39"))
	styleDimmed   = styleCard.Faint(true)
	styleHint     = lipgloss.NewStyle().Faint(true)
	styleWarn     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	stylePanel    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(0, 1)
	styleDone     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleChip     = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
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
	if m.detailOpen {
		body = m.composeDetail(body)
	}
	// Window the body to the terminal height once we know it; the footer is
	// always pinned on its own line below.
	if m.height > 1 {
		body = window(body, m.height-1, m.activeScroll())
	}
	return zone.Scan(body + "\n" + m.footer())
}

// activeScroll returns the scroll offset for whichever pane the user is driving.
func (m Model) activeScroll() int {
	if m.detailOpen {
		return m.detailScrol
	}
	return m.boardScroll
}

// window clips rendered content to height rows starting at offset, so content
// taller than the terminal scrolls instead of overflowing. offset is clamped so
// the last page stays in view.
func window(s string, height, offset int) string {
	if height <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= height {
		return s
	}
	maxOffset := len(lines) - height
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	return strings.Join(lines[offset:offset+height], "\n")
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
	keys := "tab: view  ·  ←/→: select  ·  ↑/↓: scroll  ·  enter: open  ·  q: quit"
	if m.detailOpen {
		keys = "↑/↓: scroll  ·  esc: close  ·  q: quit"
	}
	return styleHint.Render(fmt.Sprintf("[%s]  %s", v, keys))
}

// selectedChange returns the workstream the cursor is on, or nil.
func (m Model) selectedChange() *ir.Change {
	if len(m.flat) > 0 && m.selected >= 0 && m.selected < len(m.flat) {
		return m.flat[m.selected]
	}
	return nil
}

// narrow reports whether the terminal is too cramped to lay columns side by
// side, in which case the board degrades to a single scrolling list.
func (m Model) narrow(numCols int) bool {
	return m.width > 0 && m.width < minColWidth*numCols
}

// kanbanView lays out lifecycle columns side by side, each card a workstream.
// When the terminal is too narrow it degrades to a single vertical list.
func (m Model) kanbanView() string {
	if m.narrow(len(LifecycleOrder)) {
		return m.listView()
	}
	colW := m.colWidth(len(LifecycleOrder))
	var cols []string
	for _, lc := range LifecycleOrder {
		items := m.columns[lc]
		header := styleHeader.Render(fmt.Sprintf("%s (%d)", lc, len(items)))
		cards := []string{header, ""}
		for _, c := range items {
			cards = append(cards, m.card(c, m.cardStyle(c)))
		}
		col := lipgloss.JoinVertical(lipgloss.Left, cards...)
		cols = append(cols, m.colStyle(colW).Render(col))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// listView is the narrow-terminal fallback: a flat, lifecycle-grouped list.
func (m Model) listView() string {
	var b strings.Builder
	for _, lc := range LifecycleOrder {
		items := m.columns[lc]
		if len(items) == 0 {
			continue
		}
		b.WriteString(styleHeader.Render(fmt.Sprintf("%s (%d)", lc, len(items))) + "\n")
		for _, c := range items {
			b.WriteString(m.card(c, m.cardStyle(c)) + "\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// colWidth divides the terminal evenly among numCols columns; returns 0 when
// the size is unknown so styles fall back to their natural width.
func (m Model) colWidth(numCols int) int {
	if m.width <= 0 || numCols <= 0 {
		return 0
	}
	return m.width / numCols
}

// colStyle returns the column container, width-constrained when sized.
func (m Model) colStyle(w int) lipgloss.Style {
	if w <= 0 {
		return styleColumn
	}
	return styleColumn.Width(w)
}

// cardStyle picks the border treatment for a card given the current selection
// and (in graph view) its relationship to the selected node.
func (m Model) cardStyle(c *ir.Change) lipgloss.Style {
	sel := m.selectedChange()
	if sel == c {
		return styleActive
	}
	if m.view == viewGraph && sel != nil {
		nb := m.neighbors(sel.Name)
		switch {
		case nb[c.Name]:
			return styleNeighbor
		default:
			return styleDimmed
		}
	}
	return styleCard
}

// card renders a single workstream with its progress, using the supplied style.
func (m Model) card(c *ir.Change, style lipgloss.Style) string {
	done, total := Progress(c)
	label := fmt.Sprintf("%s\n%d/%d tasks", c.Name, done, total)
	return zone.Mark("card:"+c.Name, style.Render(label))
}

// graphView renders the dependency DAG as depth-ordered columns. On selection
// it emphasizes the selected node's neighbors and dims the rest (focus+context),
// keeping the layered layout rather than drawing routed edges.
func (m Model) graphView() string {
	if m.graph == nil || len(m.graph.Nodes) == 0 {
		return styleHint.Render("No dependency graph. Add edges in openspec/specutil.yaml " +
			"(or run `specutil graph --suggest`).")
	}
	cols := layers(m.graph)
	if m.narrow(len(cols)) {
		return m.graphListView(cols)
	}
	colW := m.colWidth(len(cols))
	var rendered []string
	for depth, nodes := range cols {
		header := styleHeader.Render(fmt.Sprintf("depth %d", depth))
		cells := []string{header, ""}
		for _, n := range nodes {
			c := m.byName[n.ID]
			cells = append(cells, m.card(c, m.cardStyle(c)))
		}
		col := lipgloss.JoinVertical(lipgloss.Left, cells...)
		rendered = append(rendered, m.colStyle(colW).Render(col))
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

// graphListView is the narrow-terminal fallback for the graph: nodes listed by
// depth, with the edge list following.
func (m Model) graphListView(cols [][]graph.Node) string {
	var b strings.Builder
	for depth, nodes := range cols {
		b.WriteString(styleHeader.Render(fmt.Sprintf("depth %d", depth)) + "\n")
		for _, n := range nodes {
			c := m.byName[n.ID]
			b.WriteString(m.card(c, m.cardStyle(c)) + "\n")
		}
		b.WriteString("\n")
	}
	if edges := m.edgeList(); edges != "" {
		b.WriteString(styleHint.Render("edges:\n") + edges)
	}
	return strings.TrimRight(b.String(), "\n")
}

// edgeList prints the prerequisite -> dependent relations the columns imply but
// can't draw, so the dependency direction stays legible. When a node is
// selected its incident edges are emphasized and the rest dimmed.
func (m Model) edgeList() string {
	if m.graph == nil {
		return ""
	}
	sel := m.selectedChange()
	lines := make([]string, 0, len(m.graph.Edges))
	for _, e := range m.graph.Edges {
		line := fmt.Sprintf("  %s → %s", e.From, e.To)
		if sel != nil {
			if e.From == sel.Name || e.To == sel.Name {
				line = styleChip.Render(line)
			} else {
				line = styleHint.Render(line)
			}
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// dependsOn returns the prerequisites of name: edges whose target is name.
func (m Model) dependsOn(name string) []string {
	if m.graph == nil {
		return nil
	}
	var out []string
	for _, e := range m.graph.Edges {
		if e.To == name {
			out = append(out, e.From)
		}
	}
	sort.Strings(out)
	return out
}

// blocks returns the dependents of name: edges whose source is name.
func (m Model) blocks(name string) []string {
	if m.graph == nil {
		return nil
	}
	var out []string
	for _, e := range m.graph.Edges {
		if e.From == name {
			out = append(out, e.To)
		}
	}
	sort.Strings(out)
	return out
}

// neighbors is the set of immediate prerequisites and dependents of name, used
// for focus+context highlighting.
func (m Model) neighbors(name string) map[string]bool {
	nb := make(map[string]bool)
	for _, n := range m.dependsOn(name) {
		nb[n] = true
	}
	for _, n := range m.blocks(name) {
		nb[n] = true
	}
	return nb
}

// composeDetail places the ticket panel beside the board when there is room,
// stacking it below on narrow or unsized terminals.
func (m Model) composeDetail(board string) string {
	panel := m.detailPanel()
	if m.width > 0 && !m.narrow(2) {
		panelW := m.width / 3
		board = lipgloss.NewStyle().Width(m.width - panelW - 2).Render(board)
		return lipgloss.JoinHorizontal(lipgloss.Top, board, stylePanel.Width(panelW).Render(panel))
	}
	return board + "\n\n" + stylePanel.Render(panel)
}

// detailPanel renders the selected workstream as a ticket: lifecycle, progress,
// why/what-changes, tasks-by-phase with done glyphs, and depends-on/blocks.
func (m Model) detailPanel() string {
	c := m.selectedChange()
	if c == nil {
		return styleHint.Render("Nothing selected.")
	}
	done, total := Progress(c)
	var b strings.Builder

	b.WriteString(styleHeader.Render(c.Name) + "\n")
	b.WriteString(fmt.Sprintf("[%s]  %s\n\n", Classify(c), progressBar(done, total)))

	if c.Proposal != nil && c.Proposal.Why != "" {
		b.WriteString(styleHeader.Render("Why") + "\n" + c.Proposal.Why + "\n\n")
	}
	if c.Proposal != nil && c.Proposal.WhatChanges != "" {
		b.WriteString(styleHeader.Render("What changes") + "\n" + c.Proposal.WhatChanges + "\n\n")
	}

	if c.Tasks != nil && len(c.Tasks.Phases) > 0 {
		b.WriteString(styleHeader.Render("Tasks") + "\n")
		for _, p := range c.Tasks.Phases {
			b.WriteString(fmt.Sprintf("%s. %s\n", p.Number, p.Name))
			for _, it := range p.Items {
				glyph := "[ ]"
				line := fmt.Sprintf("  %s %s", glyph, it.Text)
				if it.Done {
					line = styleDone.Render(fmt.Sprintf("  [x] %s", it.Text))
				}
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
	}

	if dep := m.dependsOn(c.Name); len(dep) > 0 {
		b.WriteString("Depends on: " + styleChip.Render(strings.Join(dep, ", ")) + "\n")
	}
	if blk := m.blocks(c.Name); len(blk) > 0 {
		b.WriteString("Blocks: " + styleChip.Render(strings.Join(blk, ", ")) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// progressBar draws a compact done/total bar like "3/5 ███░░".
func progressBar(done, total int) string {
	const width = 10
	filled := 0
	if total > 0 {
		filled = done * width / total
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%d/%d %s", done, total, bar)
}
