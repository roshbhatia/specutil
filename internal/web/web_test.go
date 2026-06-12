package web

import (
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/detail"
	"github.com/roshbhatia/specutil/internal/graph"
)

func TestRenderIsSelfContained(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db", Label: "db"}, {ID: "api", Label: "api"}},
		Edges: []graph.Edge{{From: "db", To: "api"}},
	}
	d := &detail.Feed{Changes: []detail.Change{
		{Name: "db", Lifecycle: "active", Done: 1, Total: 2},
		{Name: "api", Lifecycle: "proposed", Done: 0, Total: 3},
	}}
	out, err := Render(g, d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	html := string(out)

	// The page must carry its runtime and data inline — no external requests, so
	// it works offline from file://.
	for _, want := range []string{
		"<!doctype html>",
		"cytoscape.use(cytoscapeDagre)", // inlined bootstrap
		"The Cytoscape Consortium",      // a marker from the vendored cytoscape bundle
		"const GRAPH =",                 // graph data island
		"const DETAIL =",                // detail data island
		`"db"`,                          // node from the inlined graph data
		`"api"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}

	// No <script src> / <link href> tags — the runtime is inlined, not fetched.
	// (Plain http(s) substrings are expected inside the bundle, e.g. SVG xmlns.)
	for _, bad := range []string{"<script src", "<link "} {
		if strings.Contains(html, bad) {
			t.Errorf("page references external resource %q; must be self-contained", bad)
		}
	}
}

func TestRenderEmptyState(t *testing.T) {
	out, err := Render(&graph.Graph{}, &detail.Feed{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "No workstreams to display") {
		t.Errorf("empty graph should render guidance, got:\n%s", html)
	}
}

func TestRenderEdgelessState(t *testing.T) {
	// Nodes but no edges: the page must point the user at `graph --suggest`.
	g := &graph.Graph{Nodes: []graph.Node{{ID: "solo", Label: "solo"}}}
	out, err := Render(g, &detail.Feed{Changes: []detail.Change{{Name: "solo", Lifecycle: "proposed"}}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "graph --suggest") {
		t.Errorf("edgeless page should point at `graph --suggest`, got:\n%s", html)
	}
	// The canvas (not the empty-state guidance) must still render.
	if strings.Contains(html, "No workstreams to display") {
		t.Error("a graph with nodes should not show the empty-state guidance")
	}
}

func TestRenderNilArgs(t *testing.T) {
	if _, err := Render(nil, nil); err != nil {
		t.Fatalf("Render(nil, nil) should not error: %v", err)
	}
}

func TestRenderEscapesScriptBreakout(t *testing.T) {
	// A label that tries to close the script block must be escaped so it can't
	// break out of the inlined <script> data island.
	g := &graph.Graph{Nodes: []graph.Node{{ID: "x", Label: "</script><b>"}}}
	d := &detail.Feed{Changes: []detail.Change{{Name: "</script><b>", Lifecycle: "proposed"}}}
	out, err := Render(g, d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(string(out), "</script><b>") {
		t.Error("script-breakout label was not escaped in the data island")
	}
}
