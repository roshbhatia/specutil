package web

import (
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/graph"
)

func TestRenderIsSelfContained(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{{ID: "db", Label: "db"}, {ID: "api", Label: "api"}},
		Edges: []graph.Edge{{From: "db", To: "api"}},
	}
	out, err := Render(g)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	html := string(out)

	// The page must carry its runtime and data inline — no external requests, so
	// it works offline from file://.
	for _, want := range []string{
		"<!doctype html>",
		"mermaid.initialize",    // inlined init script
		"__esbuild_esm_mermaid", // a marker from the vendored mermaid bundle
		`"db"`,                  // node label from the inlined graph data
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
	out, err := Render(&graph.Graph{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	html := string(out)
	if !strings.Contains(html, "No workstreams to display") {
		t.Errorf("empty graph should render guidance, got:\n%s", html)
	}
}

func TestRenderNilGraph(t *testing.T) {
	if _, err := Render(nil); err != nil {
		t.Fatalf("Render(nil) should not error: %v", err)
	}
}

func TestRenderEscapesScriptBreakout(t *testing.T) {
	// A label that tries to close the script block must be escaped so it can't
	// break out of the inlined <script> data island.
	g := &graph.Graph{Nodes: []graph.Node{{ID: "x", Label: "</script><b>"}}}
	out, err := Render(g)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(string(out), "</script><b>") {
		t.Error("script-breakout label was not escaped in the data island")
	}
}
