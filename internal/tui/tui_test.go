package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
)

func TestMain(m *testing.M) {
	// View() marks/scans zones; a global manager must exist or Mark panics.
	zone.NewGlobal()
	m.Run()
}

func mkChange(name string, done, total int) *ir.Change {
	items := make([]ir.TaskItem, total)
	for i := range items {
		items[i] = ir.TaskItem{ID: "1." + string(rune('1'+i)), Text: "task", Done: i < done}
	}
	return &ir.Change{Name: name, Tasks: &ir.Tasks{Phases: []ir.Phase{{Number: "1", Name: "P", Items: items}}}}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name      string
		done, tot int
		want      Lifecycle
	}{
		{"none", 0, 0, Proposed},
		{"untouched", 0, 3, Proposed},
		{"partial", 1, 3, Active},
		{"complete", 3, 3, Archived},
	}
	for _, c := range cases {
		if got := Classify(mkChange(c.name, c.done, c.tot)); got != c.want {
			t.Errorf("%s: Classify = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestKanbanViewShowsColumnsAndProgress(t *testing.T) {
	changes := []*ir.Change{
		mkChange("alpha", 0, 2), // proposed
		mkChange("beta", 1, 2),  // active
		mkChange("gamma", 2, 2), // archived
	}
	m := New(changes, &graph.Graph{}, nil)
	out := m.View()

	for _, want := range []string{"proposed", "active", "archived", "alpha", "beta", "gamma", "1/2", "2/2"} {
		if !strings.Contains(out, want) {
			t.Errorf("kanban view missing %q\n---\n%s", want, out)
		}
	}
}

func TestEmptyRepoGuidance(t *testing.T) {
	m := New(nil, nil, nil)
	out := m.View()
	if !strings.Contains(out, "No OpenSpec changes found") {
		t.Errorf("empty repo should show guidance, got:\n%s", out)
	}
}

func TestToggleToGraphView(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db", Label: "db"}, {ID: "api", Label: "api"}},
		Edges: []graph.Edge{{From: "db", To: "api"}},
	}
	m := New([]*ir.Change{mkChange("db", 0, 1), mkChange("api", 0, 1)}, g, nil)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	gm := next.(Model)
	if gm.view != viewGraph {
		t.Fatal("tab should switch to graph view")
	}
	out := gm.View()
	// The ASCII DAG must contain both node labels and a connecting edge character.
	for _, want := range []string{"db", "api", "|"} {
		if !strings.Contains(out, want) {
			t.Errorf("graph view missing %q\n---\n%s", want, out)
		}
	}
}

func TestLayersAssignsDepth(t *testing.T) {
	// db (root) -> api -> ui : a 3-deep chain plus an isolated node.
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "api"}, {ID: "db"}, {ID: "solo"}, {ID: "ui"}},
		Edges: []graph.Edge{{From: "db", To: "api"}, {From: "api", To: "ui"}},
	}
	cols := layers(g)
	if len(cols) != 3 {
		t.Fatalf("expected 3 depth columns, got %d", len(cols))
	}
	depthOf := map[string]int{}
	for d, nodes := range cols {
		for _, n := range nodes {
			depthOf[n.ID] = d
		}
	}
	if depthOf["db"] != 0 || depthOf["api"] != 1 || depthOf["ui"] != 2 {
		t.Errorf("chain depths wrong: %v", depthOf)
	}
	if depthOf["solo"] != 0 {
		t.Errorf("isolated node should be depth 0, got %d", depthOf["solo"])
	}
}

func mkRichChange(name string, done, total int) *ir.Change {
	c := mkChange(name, done, total)
	c.Proposal = &ir.Proposal{Why: "the why of " + name, WhatChanges: "what changes in " + name}
	return c
}

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func TestNarrowTerminalDegradesToList(t *testing.T) {
	changes := []*ir.Change{mkChange("alpha", 0, 2), mkChange("beta", 1, 2)}
	m := sized(New(changes, &graph.Graph{}, nil), 20, 40) // < minColWidth*3
	out := m.View()
	// Both workstreams must still be present — nothing silently dropped.
	for _, want := range []string{"alpha", "beta", "proposed", "active"} {
		if !strings.Contains(out, want) {
			t.Errorf("narrow list view missing %q\n---\n%s", want, out)
		}
	}
}

func TestWindowClipsToHeightAndScrolls(t *testing.T) {
	// Twelve lines clipped to a 3-row window: page one shows the head, scrolling
	// reveals later rows, and nothing is dropped from the underlying content.
	lines := make([]string, 12)
	for i := range lines {
		lines[i] = fmt.Sprintf("row%d", i)
	}
	full := strings.Join(lines, "\n")

	top := window(full, 3, 0)
	if !strings.Contains(top, "row0") || strings.Contains(top, "row5") {
		t.Errorf("top window should show the head only:\n%s", top)
	}
	scrolled := window(full, 3, 5)
	if !strings.Contains(scrolled, "row5") || strings.Contains(scrolled, "row0") {
		t.Errorf("scrolled window should reveal later rows:\n%s", scrolled)
	}
	// An over-scroll clamps to the last page rather than emptying the view.
	clamped := window(full, 3, 999)
	if !strings.Contains(clamped, "row11") {
		t.Errorf("over-scroll should clamp to the last page:\n%s", clamped)
	}
}

func TestDetailPanelOpenAndClose(t *testing.T) {
	changes := []*ir.Change{mkRichChange("alpha", 1, 2)}
	m := sized(New(changes, &graph.Graph{}, nil), 120, 40)

	// Enter opens the ticket with its proposal content and tasks.
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	om := opened.(Model)
	if !om.detailOpen {
		t.Fatal("enter should open the detail panel")
	}
	out := om.View()
	for _, want := range []string{"the why of alpha", "what changes in alpha", "Tasks", "1/2"} {
		if !strings.Contains(out, want) {
			t.Errorf("detail panel missing %q\n---\n%s", want, out)
		}
	}

	// Esc closes it without quitting.
	closed, cmd := om.Update(tea.KeyMsg{Type: tea.KeyEsc})
	cm := closed.(Model)
	if cm.detailOpen {
		t.Error("esc should close the detail panel")
	}
	if cmd != nil {
		t.Error("esc with a panel open should not quit")
	}
}

func TestDetailPanelShowsRelationships(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db"}, {ID: "api"}, {ID: "ui"}},
		Edges: []graph.Edge{{From: "db", To: "api"}, {From: "api", To: "ui"}},
	}
	m := New([]*ir.Change{mkChange("db", 0, 1), mkChange("api", 0, 1), mkChange("ui", 0, 1)}, g, nil)
	// api depends on db and blocks ui.
	if got := m.dependsOn("api"); len(got) != 1 || got[0] != "db" {
		t.Errorf("dependsOn(api) = %v, want [db]", got)
	}
	if got := m.blocks("api"); len(got) != 1 || got[0] != "ui" {
		t.Errorf("blocks(api) = %v, want [ui]", got)
	}
	nb := m.neighbors("api")
	if !nb["db"] || !nb["ui"] {
		t.Errorf("neighbors(api) should include db and ui, got %v", nb)
	}
}

func TestGraphFocusDimsUnrelatedNodes(t *testing.T) {
	// db -> api ; solo is unrelated. Selecting db should style api as a neighbor
	// and solo as dimmed, while keeping every node and the edge text present.
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db"}, {ID: "api"}, {ID: "solo"}},
		Edges: []graph.Edge{{From: "db", To: "api"}},
	}
	m := New([]*ir.Change{mkChange("db", 0, 1), mkChange("api", 0, 1), mkChange("solo", 0, 1)}, g, nil)
	m.view = viewGraph
	// Select db (its flat index).
	for i, c := range m.flat {
		if c.Name == "db" {
			m.selected = i
		}
	}
	// Unrelated nodes are dimmed (faint); neighbors are not.
	if m.cardStyle(m.byName["api"]).GetFaint() {
		t.Error("neighbor node should not be dimmed under focus")
	}
	if !m.cardStyle(m.byName["solo"]).GetFaint() {
		t.Error("unrelated node should be dimmed under focus")
	}
	out := m.View()
	// ASCII DAG must show all node labels; edge direction shown by '|' line.
	for _, want := range []string{"db", "api", "solo"} {
		if !strings.Contains(out, want) {
			t.Errorf("focused graph view missing %q\n---\n%s", want, out)
		}
	}
}

func TestWorkStatus(t *testing.T) {
	// Graph: db -> api -> ui ; solo has no edges
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db"}, {ID: "api"}, {ID: "ui"}, {ID: "solo"}},
		Edges: []graph.Edge{{From: "db", To: "api"}, {From: "api", To: "ui"}},
	}
	// db=proposed/no deps -> ready; api=proposed/dep on db (not done) -> waiting
	// ui=proposed/dep on api (not done) -> waiting; solo=proposed/no deps -> ready
	// archived -> done
	db := mkChange("db", 0, 1)     // proposed, no deps
	api := mkChange("api", 0, 1)   // proposed, depends on db (not done)
	ui := mkChange("ui", 0, 1)     // proposed, depends on api (not done)
	solo := mkChange("solo", 0, 1) // proposed, no edges
	done := mkChange("done", 1, 1) // archived

	m := New([]*ir.Change{db, api, ui, solo, done}, g, nil)

	cases := []struct{ name, want string }{
		{"db", "ready"},
		{"api", "waiting"},
		{"ui", "waiting"},
		{"solo", "ready"},
		{"done", "done"},
	}
	for _, tc := range cases {
		if got := m.workStatus(m.byName[tc.name]); got != tc.want {
			t.Errorf("workStatus(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}

	// active + unmet dep -> blocked; active + all deps done -> in-progress
	apiActive := mkChange("api-active", 1, 2) // active (partial)
	dbDone := mkChange("db-done", 1, 1)       // archived
	g2 := &graph.Graph{
		Nodes: []graph.Node{{ID: "db-done"}, {ID: "api-active"}},
		Edges: []graph.Edge{{From: "db-done", To: "api-active"}},
	}
	m2 := New([]*ir.Change{dbDone, apiActive}, g2, nil)
	if got := m2.workStatus(apiActive); got != "in-progress" {
		t.Errorf("active with done dep: workStatus = %q, want in-progress", got)
	}

	// active + unmet dep (proposed prereq)
	prereq := mkChange("prereq", 0, 1)     // proposed
	dep := mkChange("dep", 1, 2)           // active
	g3 := &graph.Graph{
		Nodes: []graph.Node{{ID: "prereq"}, {ID: "dep"}},
		Edges: []graph.Edge{{From: "prereq", To: "dep"}},
	}
	m3 := New([]*ir.Change{prereq, dep}, g3, nil)
	if got := m3.workStatus(dep); got != "blocked" {
		t.Errorf("active with unmet dep: workStatus = %q, want blocked", got)
	}
}

func TestDepContext(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		// a -> b -> c ; a -> d
		Edges: []graph.Edge{{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "a", To: "d"}},
	}
	// a=proposed/no deps=ready; b,d=proposed/dep on a=waiting; c=proposed/dep on b=waiting
	a := mkChange("a", 0, 1)
	b := mkChange("b", 0, 1)
	c := mkChange("c", 0, 1)
	d := mkChange("d", 0, 1)
	m := New([]*ir.Change{a, b, c, d}, g, nil)

	// "a" is ready, blocks b and d -> "blocks 2"
	if ctx := m.depContext("a", "ready"); ctx != "blocks 2" {
		t.Errorf("depContext(a, ready) = %q, want %q", ctx, "blocks 2")
	}
	// "b" is waiting (dep on a, not done), blocked by a -> "blocked by a"
	if ctx := m.depContext("b", "waiting"); ctx != "blocked by a" {
		t.Errorf("depContext(b, waiting) = %q, want %q", ctx, "blocked by a")
	}
	// "c" is waiting, blocked by b -> "blocked by b"
	if ctx := m.depContext("c", "waiting"); ctx != "blocked by b" {
		t.Errorf("depContext(c, waiting) = %q, want %q", ctx, "blocked by b")
	}
	// no dep context for something with nothing to say (no edges at all)
	solo := mkChange("solo", 0, 1)
	mSolo := New([]*ir.Change{solo}, &graph.Graph{}, nil)
	if ctx := mSolo.depContext("solo", "ready"); ctx != "" {
		t.Errorf("depContext(solo, ready) = %q, want empty", ctx)
	}
}

func TestCardShowsStatusDotAndDepContext(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "prereq"}, {ID: "dep"}},
		Edges: []graph.Edge{{From: "prereq", To: "dep"}},
	}
	prereq := mkChange("prereq", 0, 1)  // proposed, no deps -> ready
	dep := mkChange("dep", 1, 2)        // active, unmet dep -> blocked

	m := New([]*ir.Change{prereq, dep}, g, nil)

	// Card for ready prereq should have dot and "blocks dep"
	prereqCard := m.card(prereq, styleCard)
	if !strings.Contains(prereqCard, "●") {
		t.Error("card missing status dot ●")
	}
	if !strings.Contains(prereqCard, "blocks dep") {
		t.Errorf("ready card should show 'blocks dep', got:\n%s", prereqCard)
	}

	// Card for blocked dep should have "blocked by prereq"
	depCard := m.card(dep, styleCard)
	if !strings.Contains(depCard, "blocked by prereq") {
		t.Errorf("blocked card should show 'blocked by prereq', got:\n%s", depCard)
	}
}

func TestLayersToleratesCycle(t *testing.T) {
	// a <-> b cycle must not loop forever; both pin to depth 0.
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "a"}, {ID: "b"}},
		Edges: []graph.Edge{{From: "a", To: "b"}, {From: "b", To: "a"}},
	}
	cols := layers(g) // must terminate
	if len(cols) == 0 {
		t.Fatal("expected at least one column")
	}
}
