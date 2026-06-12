package tui

import "github.com/roshbhatia/specutil/internal/graph"

// layers groups graph nodes into dependency-depth columns: depth 0 holds nodes
// with no prerequisites, and each node sits one level past its deepest
// prerequisite. This is the layered-columns layout the design chose over
// free-form edge routing (terminal DAG layout is hard). Nodes in a cycle, which
// have no well-defined depth, are pinned to depth 0; the caller surfaces the
// cycle diagnostic separately rather than looping forever.
func layers(g *graph.Graph) [][]graph.Node {
	prereqs := make(map[string][]string)
	for _, e := range g.Edges {
		prereqs[e.To] = append(prereqs[e.To], e.From)
	}

	depth := make(map[string]int, len(g.Nodes))
	var resolve func(id string, seen map[string]bool) int
	resolve = func(id string, seen map[string]bool) int {
		if d, ok := depth[id]; ok {
			return d
		}
		if seen[id] {
			return 0 // cycle: stop recursing, pin to depth 0
		}
		seen[id] = true
		best := 0
		for _, p := range prereqs[id] {
			if d := resolve(p, seen) + 1; d > best {
				best = d
			}
		}
		delete(seen, id)
		depth[id] = best
		return best
	}

	maxDepth := 0
	for _, n := range g.Nodes {
		if d := resolve(n.ID, map[string]bool{}); d > maxDepth {
			maxDepth = d
		}
	}

	cols := make([][]graph.Node, maxDepth+1)
	for _, n := range g.Nodes { // g.Nodes is already sorted, so columns are stable
		cols[depth[n.ID]] = append(cols[depth[n.ID]], n)
	}
	return cols
}
