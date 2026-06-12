// Package web renders the cross-change dependency DAG into a single
// self-contained HTML file: Cytoscape.js (plus dagre layout), the graph feed,
// and the per-workstream detail feed are all inlined so the page works offline
// from file:// with no server and no external requests. This keeps the binary
// within the determinism boundary — it writes a static artifact and never opens
// a socket.
package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/roshbhatia/specutil/internal/detail"
	"github.com/roshbhatia/specutil/internal/graph"
)

//go:embed assets/cytoscape.min.js assets/dagre.min.js assets/cytoscape-dagre.min.js assets/system.css assets/page.html.tmpl
var assets embed.FS

// page is the inlined data the template needs. text/template performs no
// contextual escaping, so these are emitted verbatim — safe here because the
// JS bundles are trusted vendored assets and json.Marshal already escapes
// <, >, & in the data literals.
type page struct {
	GraphJSON        string // the graph.json feed, embedded as a JS literal
	DetailJSON       string // the detail.json feed, embedded as a JS literal
	CytoscapeJS      string // the vendored Cytoscape runtime, inlined verbatim
	DagreJS          string // the vendored dagre layout engine
	CytoscapeDagreJS string // the cytoscape-dagre layout adapter
	SystemCSS        string // the vendored system.css (Mac OS theme), fonts inlined
	Empty            bool   // true when there are no nodes to draw
}

// Render returns a self-contained HTML document visualizing g, drilling into the
// detail feed d for per-workstream ticket content. Both feeds use the same
// renderer-independent schemas as graph.json / detail.json, so the data contract
// is shared with every other consumer. d may be nil.
func Render(g *graph.Graph, d *detail.Feed) ([]byte, error) {
	if g == nil {
		g = &graph.Graph{Nodes: []graph.Node{}, Edges: []graph.Edge{}}
	}
	if d == nil {
		d = &detail.Feed{Changes: []detail.Change{}}
	}
	// json.Marshal escapes <, >, & by default, so the literals are safe to inline
	// inside a <script> block without breaking out of it.
	graphData, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("encoding graph: %w", err)
	}
	detailData, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("encoding detail: %w", err)
	}

	cytoscapeJS, err := assets.ReadFile("assets/cytoscape.min.js")
	if err != nil {
		return nil, fmt.Errorf("reading embedded cytoscape: %w", err)
	}
	dagreJS, err := assets.ReadFile("assets/dagre.min.js")
	if err != nil {
		return nil, fmt.Errorf("reading embedded dagre: %w", err)
	}
	cytoscapeDagreJS, err := assets.ReadFile("assets/cytoscape-dagre.min.js")
	if err != nil {
		return nil, fmt.Errorf("reading embedded cytoscape-dagre: %w", err)
	}
	systemCSS, err := assets.ReadFile("assets/system.css")
	if err != nil {
		return nil, fmt.Errorf("reading embedded system.css: %w", err)
	}
	tmplSrc, err := assets.ReadFile("assets/page.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("reading embedded template: %w", err)
	}
	tmpl, err := template.New("page").Parse(string(tmplSrc))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, page{
		GraphJSON:        string(graphData),
		DetailJSON:       string(detailData),
		CytoscapeJS:      string(cytoscapeJS),
		DagreJS:          string(dagreJS),
		CytoscapeDagreJS: string(cytoscapeDagreJS),
		SystemCSS:        string(systemCSS),
		Empty:            len(g.Nodes) == 0,
	})
	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}
