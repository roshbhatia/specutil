package tui

import (
	"sort"
	"strings"

	ggl "github.com/gverger/go-graph-layout/layout"
)

// dagASCII renders the dependency DAG as a plain ASCII string using the
// Sugiyama layered-graph algorithm. It returns a string ready for bubbletea
// View(). When the graph is nil/empty it returns a short hint instead.
func (m Model) dagASCII() string {
	g := m.graph
	if g == nil || len(g.Nodes) == 0 {
		return styleHint.Render("No dependency graph. Add edges in openspec/specutil.yaml\n(or run `specutil graph --suggest`).")
	}

	// Stable ordering — g.Nodes is already sorted by Build(), but we re-sort
	// defensively so the layout is deterministic regardless of call order.
	ids := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		ids[i] = n.ID
	}
	sort.Strings(ids)

	strToIdx := make(map[string]uint64, len(ids))
	idxToStr := make(map[uint64]string, len(ids))
	for i, id := range ids {
		strToIdx[id] = uint64(i)
		idxToStr[uint64(i)] = id
	}

	// Build layout graph. W/H are in character cells; the layout engine
	// treats them as its coordinate unit, so node size directly controls spacing.
	const nodeH = 3
	gl := ggl.Graph{
		Nodes: make(map[uint64]ggl.Node, len(ids)),
		Edges: make(map[[2]uint64]ggl.Edge, len(g.Edges)),
	}
	for _, id := range ids {
		idx := strToIdx[id]
		label := id
		if len(label) > 20 {
			label = label[:17] + "..."
		}
		gl.Nodes[idx] = ggl.Node{W: len(label) + 4, H: nodeH}
	}
	for _, e := range g.Edges {
		from, to := strToIdx[e.From], strToIdx[e.To]
		gl.Edges[[2]uint64{from, to}] = ggl.Edge{}
	}

	// Apply Sugiyama layered layout. Delta/margin values are in character cells.
	ggl.SugiyamaLayersStrategyGraphLayout{
		CycleRemover:   ggl.NewSimpleCycleRemover(),
		LevelsAssigner: ggl.NewLayeredGraph,
		OrderingAssigner: ggl.WarfieldOrderingOptimizer{
			Epochs:                   50,
			LayerOrderingInitializer: ggl.BFSOrderingInitializer{},
			LayerOrderingOptimizer: ggl.CompositeLayerOrderingOptimizer{
				Optimizers: []ggl.LayerOrderingOptimizer{
					ggl.WMedianOrderingOptimizer{},
					ggl.SwitchAdjacentOrderingOptimizer{},
				},
			},
		}.Optimize,
		NodesHorizontalCoordinatesAssigner: ggl.BrandesKopfLayersNodesHorizontalAssigner{
			Delta: 5,
		},
		NodesVerticalCoordinatesAssigner: ggl.BasicNodesVerticalCoordinatesAssigner{
			MarginLayers:   2,
			FakeNodeHeight: 1,
		},
		EdgePathAssigner: ggl.StraightEdgePathAssigner{}.UpdateGraphLayout,
	}.UpdateGraphLayout(gl)

	// Compute bounding box and allocate rune grid.
	minX, minY, maxX, maxY := gl.BoundingBox()
	const margin = 1
	gridW := maxX - minX + 1 + margin*2
	gridH := maxY - minY + 1 + margin*2
	if gridW <= 0 || gridH <= 0 {
		return styleHint.Render("Empty graph layout.")
	}

	grid := make([][]rune, gridH)
	for i := range grid {
		grid[i] = make([]rune, gridW)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	set := func(x, y int, r rune) {
		gx := x - minX + margin
		gy := y - minY + margin
		if gx >= 0 && gx < gridW && gy >= 0 && gy < gridH {
			grid[gy][gx] = r
		}
	}

	// Draw edge paths first so nodes paint over them.
	for eid, e := range gl.Edges {
		from := eid[0]
		to := eid[1]
		fn := gl.Nodes[from]
		tn := gl.Nodes[to]
		path := e.Path
		if len(path) == 0 {
			// fallback: center of source to center of target
			path = []ggl.Position{
				{X: fn.X + fn.W/2, Y: fn.Y + fn.H/2},
				{X: tn.X + tn.W/2, Y: tn.Y + tn.H/2},
			}
		}
		for i := 1; i < len(path); i++ {
			p0, p1 := path[i-1], path[i]
			drawSegment(set, p0.X, p0.Y, p1.X, p1.Y)
		}
		// Arrow tip at final waypoint.
		if len(path) >= 2 {
			last := path[len(path)-1]
			prev := path[len(path)-2]
			switch {
			case prev.X == last.X && last.Y > prev.Y:
				set(last.X, last.Y, 'v')
			case prev.X == last.X && last.Y < prev.Y:
				set(last.X, last.Y, '^')
			case prev.Y == last.Y && last.X > prev.X:
				set(last.X, last.Y, '>')
			case prev.Y == last.Y && last.X < prev.X:
				set(last.X, last.Y, '<')
			}
		}
	}

	// Draw nodes over the edge grid.
	sel := m.selectedChange()
	for idx, node := range gl.Nodes {
		id := idxToStr[idx]
		label := id
		maxLabelW := node.W - 4
		if maxLabelW < 0 {
			maxLabelW = 0
		}
		if len(label) > maxLabelW {
			label = label[:maxLabelW]
		}
		x, y, w, h := node.X, node.Y, node.W, node.H

		// Top border
		set(x, y, '+')
		for dx := 1; dx < w-1; dx++ {
			set(x+dx, y, '-')
		}
		set(x+w-1, y, '+')

		// Side borders and interior
		for dy := 1; dy < h-1; dy++ {
			set(x, y+dy, '|')
			for dx := 1; dx < w-1; dx++ {
				set(x+dx, y+dy, ' ')
			}
			set(x+w-1, y+dy, '|')
		}

		// Label in middle row
		labelY := y + h/2
		for ci, ch := range label {
			if 2+ci >= w-2 {
				break
			}
			set(x+2+ci, labelY, ch)
		}

		// Bottom border
		set(x, y+h-1, '+')
		for dx := 1; dx < w-1; dx++ {
			set(x+dx, y+h-1, '-')
		}
		set(x+w-1, y+h-1, '+')

		// Selection marker: replace top-left corner with '*'
		if sel != nil && sel.Name == id {
			set(x, y, '*')
		}
	}

	// Render grid to string, stripping trailing spaces per line.
	var sb strings.Builder
	for _, row := range grid {
		line := strings.TrimRight(string(row), " ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// drawSegment draws a straight line segment from (x0,y0) to (x1,y1) on the
// grid. Vertical segments use '|', horizontal use '-', corners get '+'.
func drawSegment(set func(x, y int, r rune), x0, y0, x1, y1 int) {
	switch {
	case x0 == x1:
		// Vertical
		step := 1
		if y1 < y0 {
			step = -1
		}
		for y := y0; y != y1; y += step {
			set(x0, y, '|')
		}
	case y0 == y1:
		// Horizontal
		step := 1
		if x1 < x0 {
			step = -1
		}
		for x := x0; x != x1; x += step {
			set(x, y0, '-')
		}
	default:
		// L-shaped: horizontal then vertical
		step := 1
		if x1 < x0 {
			step = -1
		}
		for x := x0; x != x1; x += step {
			set(x, y0, '-')
		}
		set(x1, y0, '+')
		vstep := 1
		if y1 < y0 {
			vstep = -1
		}
		for y := y0; y != y1; y += vstep {
			set(x1, y, '|')
		}
	}
}
