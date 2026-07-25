package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// ManifestFile is the repo-relative path of the cross-change dependency
// manifest. It is deliberately separate from OpenSpec's .openspec.yaml so the
// dependency model stays framework-agnostic.
const ManifestFile = "openspec/specutil.yaml"

// Manifest is the hand-editable, repo-level dependency DAG. It accepts two
// equivalent spellings of the same edge set, because both appear in the wild:
//
//   - changes.<name>.depends_on — each change lists its prerequisites, so
//     `add-auth.depends_on: [add-db]` yields edge add-db -> add-auth.
//   - edges — an explicit from/to list, where from is the prerequisite.
//
// Both are merged and deduplicated by edges().
type Manifest struct {
	Changes   map[string]ManifestEntry `yaml:"changes"`
	Edges     []Edge                   `yaml:"edges"`
	Providers []ProviderConfig         `yaml:"providers"`
}

// ManifestEntry is one change's manifest record.
type ManifestEntry struct {
	DependsOn []string `yaml:"depends_on"`
}

// ProviderConfig declares a user-defined script adapter. The script is executed
// with {change} substituted by the --change value; its stdout is parsed as
// openspec-compatible markdown.
type ProviderConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// LoadManifest reads <repoRoot>/openspec/specutil.yaml. An absent file is not an
// error — it yields an empty manifest so a repo with no declared dependencies
// still produces a valid (edgeless) graph.
func LoadManifest(repoRoot string) (*Manifest, error) {
	path := filepath.Join(repoRoot, ManifestFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &m, nil
}

// edges flattens both manifest spellings into a deterministic, deduplicated
// directed edge list (prerequisite -> dependent).
func (m *Manifest) edges() []Edge {
	if m == nil {
		return nil
	}
	names := make([]string, 0, len(m.Changes))
	for name := range m.Changes {
		names = append(names, name)
	}
	sort.Strings(names)

	seen := make(map[Edge]bool)
	var edges []Edge
	add := func(e Edge) {
		if e.From == "" || e.To == "" || seen[e] {
			return
		}
		seen[e] = true
		edges = append(edges, e)
	}
	for _, dependent := range names {
		deps := append([]string(nil), m.Changes[dependent].DependsOn...)
		sort.Strings(deps)
		for _, prereq := range deps {
			add(Edge{From: prereq, To: dependent})
		}
	}
	for _, e := range m.Edges {
		add(e)
	}
	return edges
}
