// Package web renders the cross-change dependency DAG and per-workstream detail
// into a single static HTML file. The two data feeds and a pre-rendered inline
// SVG of the cross-change DAG are baked into the page; styling (Pico CSS) and
// the per-phase progress chart (Chart.js) load at view time from a pinned,
// SRI-protected CDN. The binary itself performs zero network I/O — it only
// writes this static artifact — so it stays within the determinism boundary;
// it is the rendered *page* that fetches its presentation layer when opened.
package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/detail"
	"github.com/roshbhatia/specutil/internal/graph"
)

//go:embed assets/page.html.tmpl
var assets embed.FS

// page is the inlined data the template needs. text/template performs no
// contextual escaping, so these are emitted verbatim — safe here because
// json.Marshal already escapes <, >, & in the data literals, and DagSVG is
// assembled from html.EscapeString'd labels below.
type page struct {
	GraphJSON   string // the graph.json feed, embedded as a JS literal
	DetailJSON  string // the detail.json feed, embedded as a JS literal
	DiagJSON    string // manifest diagnostics, embedded as a JS literal (may be [])
	SuggestJSON string // graph --suggest candidates, embedded as a JS literal (may be [])
	DagSVG      string // inline cross-change DAG; empty unless 2+ changes have edges
}

// Render returns a self-contained HTML document visualizing g, drilling into the
// detail feed d for per-workstream ticket content. Both feeds use the same
// renderer-independent schemas as graph.json / detail.json, so the data contract
// is shared with every other consumer. diags surfaces manifest problems (cycles,
// dangling references) in a health banner so a broken manifest is visible rather
// than discarded. candidates are the auto-inferred suggest edges shown in the UI
// so users don't have to run graph --suggest manually. d may be nil; diags and
// candidates may be empty.
func Render(g *graph.Graph, d *detail.Feed, diags []graph.Diagnostic, candidates []graph.Candidate) ([]byte, error) {
	if g == nil {
		g = &graph.Graph{Nodes: []graph.Node{}, Edges: []graph.Edge{}}
	}
	if d == nil {
		d = &detail.Feed{Changes: []detail.Change{}}
	}
	if diags == nil {
		diags = []graph.Diagnostic{}
	}
	if candidates == nil {
		candidates = []graph.Candidate{}
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
	diagData, err := json.Marshal(diags)
	if err != nil {
		return nil, fmt.Errorf("encoding diagnostics: %w", err)
	}
	suggestData, err := json.Marshal(candidates)
	if err != nil {
		return nil, fmt.Errorf("encoding suggestions: %w", err)
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
		GraphJSON:   string(graphData),
		DetailJSON:  string(detailData),
		DiagJSON:    string(diagData),
		SuggestJSON: string(suggestData),
		DagSVG:      dagSVG(g, lifecycleByName(d)),
	})
	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}
	return buf.Bytes(), nil
}

// lifecycleByName maps each change name to its lifecycle so the DAG can stamp
// every node with a lifecycle CSS class (the SVG owns no colors itself; CSS does,
// which is what lets the graph track the active light/dark theme).
func lifecycleByName(d *detail.Feed) map[string]string {
	m := make(map[string]string, len(d.Changes))
	for _, c := range d.Changes {
		m[c.Name] = c.Lifecycle
	}
	return m
}

// dagSVG renders the cross-change dependency DAG as a deterministic, dependency-
// free inline SVG: nodes are laid out in left-to-right columns by longest-path
// depth (so prerequisites sit left of dependents), siblings stacked by name.
// Each node is stamped with its lifecycle CSS class and wrapped in an anchor to
// its change document, so the graph is themeable (CSS owns the palette) and
// navigable (pure anchors, no script, file://-safe). lc maps change name to
// lifecycle; a name absent from lc falls back to no lifecycle class. Returns ""
// when there is nothing worth drawing (fewer than 2 nodes or no edges) — the
// overview then shows a "no dependencies" footnote instead.
func dagSVG(g *graph.Graph, lc map[string]string) string {
	if g == nil || len(g.Edges) == 0 || len(g.Nodes) < 2 {
		return ""
	}

	label := make(map[string]string, len(g.Nodes))
	ids := make([]string, 0, len(g.Nodes))
	adj := make(map[string][]string)
	indeg := make(map[string]int)
	for _, n := range g.Nodes {
		l := n.Label
		if l == "" {
			l = n.ID
		}
		label[n.ID] = l
		ids = append(ids, n.ID)
		if _, ok := indeg[n.ID]; !ok {
			indeg[n.ID] = 0
		}
	}
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
		indeg[e.To]++
	}
	sort.Strings(ids)

	// Longest-path layering via Kahn's algorithm; ties broken by name for
	// determinism. Nodes left unprocessed by a cycle keep depth 0 and still draw.
	depth := make(map[string]int)
	remaining := make(map[string]int, len(indeg))
	queue := []string{}
	for _, id := range ids {
		remaining[id] = indeg[id]
		if indeg[id] == 0 {
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		sort.Strings(queue)
		n := queue[0]
		queue = queue[1:]
		nbrs := append([]string(nil), adj[n]...)
		sort.Strings(nbrs)
		for _, to := range nbrs {
			if depth[n]+1 > depth[to] {
				depth[to] = depth[n] + 1
			}
			remaining[to]--
			if remaining[to] == 0 {
				queue = append(queue, to)
			}
		}
	}

	// Bucket nodes into columns by depth.
	maxDepth := 0
	for _, id := range ids {
		if depth[id] > maxDepth {
			maxDepth = depth[id]
		}
	}
	cols := make([][]string, maxDepth+1)
	for _, id := range ids {
		cols[depth[id]] = append(cols[depth[id]], id)
	}

	const (
		colGap = 230
		rowGap = 74
		boxW   = 190
		boxH   = 46
		padX   = 16
		padY   = 36
	)

	// Center each node and remember its anchor points for edge routing.
	type pt struct{ x, y float64 }
	left := make(map[string]pt)
	right := make(map[string]pt)
	var rects strings.Builder
	maxRows := 0
	for d, col := range cols {
		if len(col) > maxRows {
			maxRows = len(col)
		}
		rects.WriteString(fmt.Sprintf(
			`<text class="depth-label" x="%.0f" y="16">depth %d</text>`,
			float64(padX+d*colGap), d,
		))
		for i, id := range col {
			x := float64(padX + d*colGap)
			y := float64(padY + i*rowGap)
			left[id] = pt{x, y + boxH/2}
			right[id] = pt{x + boxW, y + boxH/2}
			cls := "node"
			if l := lc[id]; l != "" {
				cls += " lc-" + l
			}
			// The node is an <a> to its change document: pure-anchor navigation,
			// no script, file://-safe. data-node lets the optional hover-emphasis
			// enhancement find a node; colors come from CSS via the lifecycle class.
			rects.WriteString(fmt.Sprintf(
				`<a class="%s" data-node="%s" href="#/c/%s" aria-label="%s"><title>%s</title><rect class="nbox" x="%.0f" y="%.0f" width="%d" height="%d"/>`+
					`<text class="nlabel" x="%.0f" y="%.0f" font-size="12" text-anchor="middle" dominant-baseline="middle">%s</text></a>`,
				cls, html.EscapeString(id), urlHashEscape(id), html.EscapeString(label[id]), html.EscapeString(label[id]),
				x, y, boxW, boxH, x+boxW/2, y+boxH/2, html.EscapeString(truncate(label[id], 24)),
			))
		}
	}

	// Edges: prerequisite (From, left column) -> dependent (To, right column).
	var lines strings.Builder
	edges := append([]graph.Edge(nil), g.Edges...)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	for _, e := range edges {
		from, okF := right[e.From]
		to, okT := left[e.To]
		if !okF || !okT {
			continue
		}
		lines.WriteString(fmt.Sprintf(
			`<line class="edge" data-from="%s" data-to="%s" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke-width="1.5" marker-end="url(#arrow)"/>`,
			html.EscapeString(e.From), html.EscapeString(e.To), from.x, from.y, to.x, to.y,
		))
	}

	w := padX*2 + (maxDepth+1)*colGap - (colGap - boxW)
	h := padY*2 + maxRows*rowGap - (rowGap - boxH)
	// The marker path and edges carry no inline fill/stroke; CSS colors them from
	// theme variables so the arrowheads track light/dark like everything else.
	return fmt.Sprintf(
		`<svg viewBox="0 0 %d %d" width="%d" height="%d" role="img" aria-label="Cross-change dependency graph">`+
			`<defs><marker id="arrow" class="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">`+
			`<path d="M0,0 L10,5 L0,10 z"/></marker></defs>%s%s</svg>`,
		w, h, w, h, lines.String(), rects.String(),
	)
}

// urlHashEscape escapes a change name for safe inclusion in a #/c/<name> hash
// route, mirroring the JS encodeURIComponent used elsewhere for the same links.
func urlHashEscape(s string) string {
	return strings.NewReplacer(
		"%", "%25", "#", "%23", "?", "%3F", "&", "%26", " ", "%20",
		`"`, "%22", "<", "%3C", ">", "%3E",
	).Replace(s)
}

// truncate clips a label to n runes, appending an ellipsis when shortened, so
// long change names don't overflow their fixed-width DAG box.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 1 {
		return ""
	}
	return string(r[:n-1]) + "…"
}
