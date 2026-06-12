// Package web renders the cross-change dependency DAG into a single
// self-contained HTML file: Mermaid and the graph data are inlined so the page
// works offline from file:// with no server and no external requests. This
// keeps the binary within the determinism boundary — it writes a static
// artifact and never opens a socket.
package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/roshbhatia/specutil/internal/graph"
)

//go:embed assets/mermaid.min.js assets/page.html.tmpl
var assets embed.FS

// page is the inlined data the template needs. text/template performs no
// contextual escaping, so these are emitted verbatim — safe here because the
// Mermaid bundle is a trusted vendored asset and json.Marshal already escapes
// <, >, & in the graph literal.
type page struct {
	GraphJSON string // the graph.json feed, embedded as a JS literal
	MermaidJS string // the vendored Mermaid runtime, inlined verbatim
	Empty     bool   // true when there are no nodes to draw
}

// Render returns a self-contained HTML document visualizing g. The page reads
// the embedded graph data (the same renderer-independent graph.json schema) and
// builds the Mermaid diagram client-side, so swapping renderers later needs no
// change to the data contract.
func Render(g *graph.Graph) ([]byte, error) {
	if g == nil {
		g = &graph.Graph{Nodes: []graph.Node{}, Edges: []graph.Edge{}}
	}
	// json.Marshal escapes <, >, & by default, so the literal is safe to inline
	// inside a <script> block without breaking out of it.
	data, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("encoding graph: %w", err)
	}
	mermaidJS, err := assets.ReadFile("assets/mermaid.min.js")
	if err != nil {
		return nil, fmt.Errorf("reading embedded mermaid: %w", err)
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
		GraphJSON: string(data),
		MermaidJS: string(mermaidJS),
		Empty:     len(g.Nodes) == 0,
	})
	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}
