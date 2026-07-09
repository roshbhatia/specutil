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
		case "g":
			if m.detailOpen {
				m.detailScrol = 0
			} else {
				m.boardScroll = 0
			}
		case "G":
			if m.detailOpen {
				m.detailScrol = 9999
			} else {
				m.boardScroll = 9999
			}
		case " ":
			pageSize := m.height / 3
			if pageSize < 1 {
				pageSize = 5
			}
			m.scroll(pageSize)
		case "b":
			pageSize := m.height / 3
			if pageSize < 1 {
				pageSize = 5
			}
			m.scroll(-pageSize)
		case "ctrl+d":
			half := m.height / 2
			if half < 3 {
				half = 3
			}
			m.scroll(half)
		case "ctrl+u":
			half := m.height / 2
			if half < 3 {
				half = 3
			}
			m.scroll(-half)
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.scroll(-3)
		case tea.MouseButtonWheelDown:
			m.scroll(3)
		default:
			if msg.Action == tea.MouseActionPress {
				for i, c := range m.flat {
					if zone.Get("card:"+c.Name).InBounds(msg) {
						m.selected = i
						m.detailOpen = true
						m.detailScrol = 0
					}
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

// Palette. Every color is an AdaptiveColor so the TUI tracks the terminal's
// light/dark background (lipgloss resolves the variant from the detected
// background) the same way the web viewer follows prefers-color-scheme — the two
// surfaces share one visual language. Lifecycle hues mirror the web's badges:
// slate=proposed, amber=active, green=archived.
var (
	colProposed = lipgloss.AdaptiveColor{Light: "245", Dark: "245"} // slate/gray
	colActive   = lipgloss.AdaptiveColor{Light: "172", Dark: "214"} // amber
	colArchived = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}   // green
	colAccent   = lipgloss.AdaptiveColor{Light: "27", Dark: "39"}   // selection blue
	colNeighbor = lipgloss.AdaptiveColor{Light: "31", Dark: "45"}   // neighbor cyan
	colDone     = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}   // completed green
	colWarn     = lipgloss.AdaptiveColor{Light: "160", Dark: "203"} // diagnostics red
	colMuted    = lipgloss.AdaptiveColor{Light: "242", Dark: "245"} // secondary text
	colTrack    = lipgloss.AdaptiveColor{Light: "252", Dark: "238"} // progress track
)

var (
	styleColumn   = lipgloss.NewStyle().Padding(0, 1)
	styleHeader   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleCard     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Margin(0, 0, 1, 0)
	styleActive   = styleCard.BorderForeground(colAccent)
	styleNeighbor = styleCard.BorderForeground(colNeighbor)
	styleDimmed   = styleCard.Faint(true)
	styleHint     = lipgloss.NewStyle().Faint(true)
	styleMuted    = lipgloss.NewStyle().Foreground(colMuted)
	styleWarn     = lipgloss.NewStyle().Foreground(colWarn)
	stylePanel    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colAccent).Padding(0, 1)
	styleDone     = lipgloss.NewStyle().Foreground(colDone)
	styleChip     = lipgloss.NewStyle().Foreground(colAccent)
	styleName     = lipgloss.NewStyle().Bold(true)

	// Task-kind markers mirror the web checklist tags; plain tasks get none.
	styleKindVerify  = lipgloss.NewStyle().Foreground(colProposed).Bold(true)
	styleKindApply   = lipgloss.NewStyle().Foreground(colActive).Bold(true)
	styleKindConfirm = lipgloss.NewStyle().Foreground(colArchived).Bold(true)
)

// lifecycleColor maps a lifecycle to its adaptive hue, shared by cards, column
// headers, and the detail badge so the state reads consistently everywhere.
func lifecycleColor(lc Lifecycle) lipgloss.AdaptiveColor {
	switch lc {
	case Active:
		return colActive
	case Archived:
		return colArchived
	default:
		return colProposed
	}
}

// lifecycleHeader is a lifecycle-colored, bold/underlined column header like
// "active (2)", keyed to the same hue the cards in that column use.
func lifecycleHeader(lc Lifecycle, n int) string {
	return styleHeader.Foreground(lifecycleColor(lc)).Render(fmt.Sprintf("%s (%d)", lc, n))
}

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
	keys := "tab: view  ·  ←/→: select  ·  j/k: scroll  ·  ^d/^u: page  ·  g/G: top/bot  ·  enter: open  ·  q: quit"
	if m.detailOpen {
		keys = "j/k: scroll  ·  ^d/^u: page  ·  g/G: top/bot  ·  esc: close  ·  q: quit"
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
		header := lifecycleHeader(lc, len(items))
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
		b.WriteString(lifecycleHeader(lc, len(items)) + "\n")
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
	// Unfocused cards carry their lifecycle hue on the border so the board reads
	// like the web's lifecycle-colored cards at a glance.
	return styleCard.BorderForeground(lifecycleColor(Classify(c)))
}

// card renders a single workstream as a compact progress card: name, an inline
// lifecycle-colored meter, done/total, and a phase count — mirroring the web
// board's cards. The pieces degrade gracefully on a narrow terminal because the
// column width clips them rather than overflowing.
func (m Model) card(c *ir.Change, style lipgloss.Style) string {
	done, total := Progress(c)
	lc := Classify(c)
	phases := 0
	if c.Tasks != nil {
		phases = len(c.Tasks.Phases)
	}
	var b strings.Builder
	b.WriteString(styleName.Render(c.Name) + "\n")
	b.WriteString(miniBar(done, total, lifecycleColor(lc)) + "\n")
	b.WriteString(styleMuted.Render(fmt.Sprintf("%d/%d · %d phase%s", done, total, phases, plural(phases))))
	return zone.Mark("card:"+c.Name, style.Render(b.String()))
}

// miniBar draws a short fixed-width progress meter; the filled run takes the
// supplied (lifecycle) color, the remainder a muted track.
func miniBar(done, total int, c lipgloss.TerminalColor) string {
	const w = 8
	filled := 0
	if total > 0 {
		filled = done * w / total
	}
	if filled > w {
		filled = w
	}
	fill := lipgloss.NewStyle().Foreground(c).Render(strings.Repeat("█", filled))
	track := lipgloss.NewStyle().Foreground(colTrack).Render(strings.Repeat("░", w-filled))
	return fill + track
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// depthHeader labels a graph column by dependency depth, naming depth 0 as the
// roots (no prerequisites) and the rest by depth, with the node count. The
// literal "depth N" stays in the string so the layout reads unambiguously.
func depthHeader(depth, n int) string {
	role := ""
	if depth == 0 {
		role = " · roots"
	}
	return styleHeader.Render(fmt.Sprintf("depth %d%s (%d)", depth, role, n))
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
		header := depthHeader(depth, len(nodes))
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
		b.WriteString(depthHeader(depth, len(nodes)) + "\n")
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

// composeDetail renders the detail panel full-screen, replacing the board.
// Esc closes it and returns to the board. Full-width avoids the cramped
// right-column layout and lets the task checklist breathe.
func (m Model) composeDetail(_ string) string {
	frame := stylePanel.GetHorizontalFrameSize()
	w := m.width - frame
	if w < 1 {
		w = 1
	}
	innerW := w - stylePanel.GetHorizontalPadding()
	if innerW < 1 {
		innerW = 1
	}
	return stylePanel.Width(w).Render(m.detailPanel(innerW))
}

// detailPanel renders the selected workstream as a ticket. innerW is the panel's
// content width: when it's wide enough the tasks-by-phase checklist sits to the
// left of a relationships/per-phase rail (matching the web document's two-column
// layout); when narrow the rail stacks *above* the checklist so relationships
// are read alongside the work rather than buried beneath a long list.
func (m Model) detailPanel(innerW int) string {
	c := m.selectedChange()
	if c == nil {
		return styleHint.Render("Nothing selected.")
	}
	done, total := Progress(c)

	var head strings.Builder
	head.WriteString(styleHeader.Render(c.Name) + "\n")
	head.WriteString(lifecycleBadge(Classify(c)) + "  " + progressBar(done, total))

	var meta strings.Builder
	if c.Proposal != nil && c.Proposal.Why != "" {
		meta.WriteString(styleHeader.Render("Why") + "\n" + c.Proposal.Why + "\n\n")
	}
	if c.Proposal != nil && c.Proposal.WhatChanges != "" {
		meta.WriteString(styleHeader.Render("What changes") + "\n" + c.Proposal.WhatChanges + "\n\n")
	}

	pipeline := m.detailPipeline(c)
	outstanding := m.detailOutstanding(c)
	checklist := m.detailChecklist(c)
	rail := m.detailRail(c)

	const sideBySideMin = 56
	var body string
	switch {
	case innerW >= sideBySideMin && checklist != "":
		railW := 22
		listW := innerW - railW - 1
		main := outstanding + "\n\n" + checklist
		listCol := lipgloss.NewStyle().Width(listW).Render(main)
		railCol := lipgloss.NewStyle().Width(railW).Render(rail)
		body = lipgloss.JoinHorizontal(lipgloss.Top, listCol, " ", railCol)
	case checklist != "":
		body = rail + "\n\n" + outstanding + "\n\n" + checklist
	default:
		body = rail
	}

	parts := []string{strings.TrimRight(head.String(), "\n")}
	if pipeline != "" {
		parts = append(parts, pipeline)
	}
	if s := strings.TrimRight(meta.String(), "\n"); s != "" {
		parts = append(parts, s)
	}
	parts = append(parts, body)
	return strings.Join(parts, "\n\n")
}

// detailPipeline renders the stage sequence compactly:
// "Stage 1: Setup → Stage 2: Implement → Stage 3: Verify"
// so the sequential/parallel structure is obvious before the full checklist.
func (m Model) detailPipeline(c *ir.Change) string {
	if c.Tasks == nil || len(c.Tasks.Phases) == 0 {
		return ""
	}
	var stgs []string
	for _, p := range c.Tasks.Phases {
		pd, pt := phaseProgress(p)
		pct := 0
		if pt > 0 {
			pct = pd * 100 / pt
		}
		label := "Stage " + p.Number + ": " + p.Name
		stgs = append(stgs, label+styleMuted.Render(fmt.Sprintf(" %d%%", pct)))
	}
	arrow := styleHint.Render(" → ")
	row := strings.Join(stgs, arrow)
	return styleHeader.Render("Execution plan") + "  " +
		styleHint.Render("(tasks run when their dependencies are met)") + "\n" + row
}

// detailOutstanding shows the first few incomplete tasks so the most urgent
// work is visible without scrolling through the full checklist.
func (m Model) detailOutstanding(c *ir.Change) string {
	if c.Tasks == nil {
		return ""
	}
	var remaining []struct {
		key  string
		text string
		kind ir.TaskKind
	}
	for pi, p := range c.Tasks.Phases {
		for ii, it := range p.Items {
			if !it.Done {
				key := fmt.Sprintf("%d%c", pi, rune('a'+ii))
				remaining = append(remaining, struct {
					key  string
					text string
					kind ir.TaskKind
				}{key, it.Text, it.Kind})
			}
		}
	}
	if len(remaining) == 0 {
		return styleHeader.Render("Outstanding") + "\n" + styleDone.Render("✓ All tasks complete")
	}
	const maxShow = 5
	var b strings.Builder
	b.WriteString(styleHeader.Render(fmt.Sprintf("Outstanding (%d)", len(remaining))) + "\n")
	shown := remaining
	if len(shown) > maxShow {
		shown = shown[:maxShow]
	}
	for _, r := range shown {
		key := styleMuted.Render("[" + r.key + "]")
		b.WriteString(fmt.Sprintf("  [ ] %s %s%s\n", key, kindMarker(r.kind), r.text))
	}
	if len(remaining) > maxShow {
		b.WriteString(styleHint.Render(fmt.Sprintf("  … and %d more", len(remaining)-maxShow)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// detailChecklist is the tasks-by-phase list with per-phase progress and
// verify/apply/confirm markers; completed items are colored done.
func (m Model) detailChecklist(c *ir.Change) string {
	if c.Tasks == nil || len(c.Tasks.Phases) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(styleHeader.Render("Tasks") + "\n")
	for _, p := range c.Tasks.Phases {
		pd, pt := phaseProgress(p)
		b.WriteString(fmt.Sprintf("Stage %s: %s %s\n", p.Number, p.Name, styleMuted.Render(fmt.Sprintf("(%d/%d)", pd, pt))))
		for _, it := range p.Items {
			glyph := "[ ]"
			if it.Done {
				glyph = "[x]"
			}
			line := fmt.Sprintf("  %s %s%s", glyph, kindMarker(it.Kind), it.Text)
			if it.Done {
				line = styleDone.Render(line)
			}
			b.WriteString(line + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// detailRail is the side rail: per-phase progress meters and the depends-on /
// blocks relationships, the TUI counterpart to the web document's aside.
func (m Model) detailRail(c *ir.Change) string {
	var b strings.Builder
	if c.Tasks != nil && len(c.Tasks.Phases) > 0 {
		b.WriteString(styleHeader.Render("Stages") + "\n")
		for _, p := range c.Tasks.Phases {
			pd, pt := phaseProgress(p)
			b.WriteString(miniBar(pd, pt, colAccent) + " " + styleMuted.Render("Stage "+p.Number+": "+p.Name) + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(styleHeader.Render("Depends on") + "\n")
	if dep := m.dependsOn(c.Name); len(dep) > 0 {
		for _, d := range dep {
			b.WriteString(styleChip.Render("• "+d) + "\n")
		}
	} else {
		b.WriteString(styleHint.Render("nothing") + "\n")
	}
	b.WriteString("\n" + styleHeader.Render("Blocks") + "\n")
	if blk := m.blocks(c.Name); len(blk) > 0 {
		for _, d := range blk {
			b.WriteString(styleChip.Render("• "+d) + "\n")
		}
	} else {
		b.WriteString(styleHint.Render("nothing") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// phaseProgress counts done/total items in a phase.
func phaseProgress(p ir.Phase) (done, total int) {
	for _, it := range p.Items {
		total++
		if it.Done {
			done++
		}
	}
	return done, total
}

// kindMarker is the inline tag for non-plain task kinds (verify/apply/confirm),
// matching the web checklist's markers. Plain tasks get nothing.
func kindMarker(k ir.TaskKind) string {
	switch k {
	case ir.KindVerify:
		return styleKindVerify.Render("verify ")
	case ir.KindApply:
		return styleKindApply.Render("apply ")
	case ir.KindConfirm:
		return styleKindConfirm.Render("confirm ")
	default:
		return ""
	}
}

// lifecycleBadge is a bold, lifecycle-colored "[active]"-style chip.
func lifecycleBadge(lc Lifecycle) string {
	return lipgloss.NewStyle().Foreground(lifecycleColor(lc)).Bold(true).Render("[" + string(lc) + "]")
}

// progressBar draws a compact done/total bar like "3/5 ███░░", the filled run
// accented so progress reads at a glance.
func progressBar(done, total int) string {
	const width = 10
	filled := 0
	if total > 0 {
		filled = done * width / total
	}
	if filled > width {
		filled = width
	}
	fill := lipgloss.NewStyle().Foreground(colAccent).Render(strings.Repeat("█", filled))
	track := lipgloss.NewStyle().Foreground(colTrack).Render(strings.Repeat("░", width-filled))
	return fmt.Sprintf("%d/%d %s%s", done, total, fill, track)
}
