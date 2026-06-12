package tui

import (
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
	for _, want := range []string{"depth 0", "depth 1", "db → api"} {
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
