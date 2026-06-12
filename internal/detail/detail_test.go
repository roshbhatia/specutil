package detail

import (
	"bytes"
	"testing"

	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/lifecycle"
)

func mkChange(name string, done, total int) *ir.Change {
	items := make([]ir.TaskItem, total)
	for i := range items {
		items[i] = ir.TaskItem{Text: "task", Done: i < done}
	}
	return &ir.Change{
		Name:     name,
		Proposal: &ir.Proposal{Why: "because", WhatChanges: "stuff"},
		Tasks:    &ir.Tasks{Phases: []ir.Phase{{Number: "1", Name: "P", Items: items}}},
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	// Input order is intentionally unsorted to prove output is stable regardless.
	changes := []*ir.Change{mkChange("gamma", 2, 2), mkChange("alpha", 0, 3), mkChange("beta", 1, 2)}

	a, err := Build(changes).JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	b, err := Build(changes).JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("detail feed not byte-identical on repeat:\n%s\n---\n%s", a, b)
	}

	// And the entries are sorted by name regardless of input order.
	f := Build(changes)
	if f.Changes[0].Name != "alpha" || f.Changes[1].Name != "beta" || f.Changes[2].Name != "gamma" {
		t.Errorf("changes not sorted by name: %v", []string{f.Changes[0].Name, f.Changes[1].Name, f.Changes[2].Name})
	}
}

func TestLifecycleParityWithSharedClassifier(t *testing.T) {
	// The detail feed must report the same lifecycle/progress the shared
	// classifier produces — the guarantee both surfaces rely on.
	cases := []*ir.Change{mkChange("p", 0, 2), mkChange("a", 1, 2), mkChange("d", 3, 3)}
	feed := Build(cases)
	byName := map[string]Change{}
	for _, c := range feed.Changes {
		byName[c.Name] = c
	}
	for _, c := range cases {
		want := string(lifecycle.Classify(c))
		wd, wt := lifecycle.Progress(c)
		got := byName[c.Name]
		if got.Lifecycle != want {
			t.Errorf("%s: lifecycle = %q, want %q", c.Name, got.Lifecycle, want)
		}
		if got.Done != wd || got.Total != wt {
			t.Errorf("%s: progress = %d/%d, want %d/%d", c.Name, got.Done, got.Total, wd, wt)
		}
	}
}

func TestItemLevelKeyTracksPhaseAndSibling(t *testing.T) {
	// Two phases: phase 1 (level 0) has two parallel items, phase 2 (level 1) one.
	c := &ir.Change{
		Name: "x",
		Tasks: &ir.Tasks{Phases: []ir.Phase{
			{Number: "1", Name: "first", Items: []ir.TaskItem{{Text: "a"}, {Text: "b"}}},
			{Number: "2", Name: "second", Items: []ir.TaskItem{{Text: "c"}}},
		}},
	}
	f := Build([]*ir.Change{c})
	ph := f.Changes[0].Phases
	if ph[0].Items[0].Level != 0 || ph[0].Items[0].Key != "0a" {
		t.Errorf("phase 1 item 1 = level %d key %q, want 0/0a", ph[0].Items[0].Level, ph[0].Items[0].Key)
	}
	if ph[0].Items[1].Level != 0 || ph[0].Items[1].Key != "0b" {
		t.Errorf("phase 1 item 2 = level %d key %q, want 0/0b", ph[0].Items[1].Level, ph[0].Items[1].Key)
	}
	if ph[1].Items[0].Level != 1 || ph[1].Items[0].Key != "1a" {
		t.Errorf("phase 2 item 1 = level %d key %q, want 1/1a", ph[1].Items[0].Level, ph[1].Items[0].Key)
	}
}

func TestBuildCarriesTaskContent(t *testing.T) {
	f := Build([]*ir.Change{mkChange("x", 1, 2)})
	c := f.Changes[0]
	if c.Why != "because" || c.WhatChanges != "stuff" {
		t.Errorf("proposal content not carried: %+v", c)
	}
	if len(c.Phases) != 1 || len(c.Phases[0].Items) != 2 {
		t.Fatalf("phases/items not carried: %+v", c.Phases)
	}
	if !c.Phases[0].Items[0].Done || c.Phases[0].Items[1].Done {
		t.Errorf("done state wrong: %+v", c.Phases[0].Items)
	}
}
