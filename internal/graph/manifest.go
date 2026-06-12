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

// Manifest is the hand-editable, repo-level dependency DAG. Dependencies are
// keyed by change name: each change lists the changes it depends on (its
// prerequisites), so `add-auth.depends_on: [add-db]` yields edge add-db ->
// add-auth.
type Manifest struct {
	Changes map[string]ManifestEntry `yaml:"changes"`
}

// ManifestEntry is one change's manifest record.
type ManifestEntry struct {
	DependsOn []string `yaml:"depends_on"`
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

// edges flattens the manifest into a deterministic directed edge list
// (prerequisite -> dependent).
func (m *Manifest) edges() []Edge {
	if m == nil {
		return nil
	}
	names := make([]string, 0, len(m.Changes))
	for name := range m.Changes {
		names = append(names, name)
	}
	sort.Strings(names)

	var edges []Edge
	for _, dependent := range names {
		deps := append([]string(nil), m.Changes[dependent].DependsOn...)
		sort.Strings(deps)
		for _, prereq := range deps {
			edges = append(edges, Edge{From: prereq, To: dependent})
		}
	}
	return edges
}
