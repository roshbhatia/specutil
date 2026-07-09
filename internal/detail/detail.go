// Package detail projects the loaded IR into detail.json: a per-change feed of
// lifecycle, progress, and task content that powers the visualizers' ticket
// drill-down. It is a renderer-independent projection alongside graph.json,
// which stays the pure dependency feed — dependsOn/blocks are derived by
// consumers from graph edges, not duplicated here. Everything is pure and
// deterministic: identical inputs yield byte-identical output.
package detail

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/lifecycle"
)

// Feed is the whole detail projection: one entry per change, sorted by name for
// deterministic output.
type Feed struct {
	Changes []Change `json:"changes"`
}

// Change is the per-workstream ticket content.
type Change struct {
	Name        string          `json:"name"`
	Lifecycle   string          `json:"lifecycle"`
	Done        int             `json:"done"`
	Total       int             `json:"total"`
	Why         string          `json:"why,omitempty"`
	WhatChanges string          `json:"whatChanges,omitempty"`
	Design      *DesignSections `json:"design,omitempty"`
	Phases      []Phase         `json:"phases"`
}

// DesignSections surfaces design.md content for visualizers.
type DesignSections struct {
	Context       string `json:"context,omitempty"`
	Goals         string `json:"goals,omitempty"`
	NonGoals      string `json:"nonGoals,omitempty"`
	Decisions     string `json:"decisions,omitempty"`
	Risks         string `json:"risks,omitempty"`
	Rollout       string `json:"rollout,omitempty"`
	Migration     string `json:"migration,omitempty"`
	OpenQuestions string `json:"openQuestions,omitempty"`
}

// Phase mirrors a tasks.md phase with its checkbox items.
type Phase struct {
	Number string `json:"number"`
	Name   string `json:"name"`
	Items  []Item `json:"items"`
}

// Item is one checkbox task. Level is the 0-based dependency rank: phases are
// sequential, so a phase's ordinal is the level, and every item in the same
// phase shares it — those items can be worked in parallel. Key disambiguates
// siblings within a level with a letter (0a, 0b, 1a, …), giving each task a
// short stable handle that reads as "what blocks what".
type Item struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
	// Kind is the verify/apply/confirm discipline classification carried from the
	// IR ("task" for plain items), so visualizers can mark impactful and
	// confirmation steps without re-parsing the source markdown.
	Kind         string        `json:"kind"`
	Level        int           `json:"level"`
	Key          string        `json:"key"`
	Tags         []string      `json:"tags,omitempty"`
	InlineRefs   []string      `json:"inlineRefs,omitempty"`
	ExternalRefs []ExternalRef `json:"externalRefs,omitempty"`
}

// ExternalRef is a confirmed mapping from a task to an external system record
// (e.g. a Linear issue or Notion page) written by `specutil lock set`.
type ExternalRef struct {
	Target     string `json:"target"`
	ExternalID string `json:"externalId"`
}

// RefsByKey maps a composite key (changeName + "\x00" + phaseName + "\x00" +
// itemText) to the external refs confirmed for that item. Built by the caller
// from the per-change lockfiles so the detail package stays free of sync deps.
type RefsByKey map[string][]ExternalRef

// levelKey renders the (level, sibling-index) pair as a compact handle: the
// 0-based level followed by a letter (a..z), falling back to the raw index past
// 26 siblings so it never collides or runs out of letters.
func levelKey(level, idx int) string {
	if idx < 26 {
		return fmt.Sprintf("%d%c", level, 'a'+idx)
	}
	return strconv.Itoa(level) + "x" + strconv.Itoa(idx)
}

// Build assembles the detail feed from the loaded changes with no external refs.
func Build(changes []*ir.Change) *Feed { return BuildWithRefs(changes, nil) }

// BuildWithRefs assembles the detail feed and annotates each task item with any
// confirmed external references from refs. refs may be nil (no-op).
func BuildWithRefs(changes []*ir.Change, refs RefsByKey) *Feed {
	out := make([]Change, 0, len(changes))
	for _, c := range changes {
		done, total := lifecycle.Progress(c)
		dc := Change{
			Name:      c.Name,
			Lifecycle: string(lifecycle.Classify(c)),
			Done:      done,
			Total:     total,
			Phases:    []Phase{},
		}
		if c.Proposal != nil {
			dc.Why = c.Proposal.Why
			dc.WhatChanges = c.Proposal.WhatChanges
		}
		if c.Design != nil {
			ds := &DesignSections{
				Context:       c.Design.Context,
				Goals:         c.Design.Goals,
				NonGoals:      c.Design.NonGoals,
				Decisions:     c.Design.Decisions,
				Risks:         c.Design.Risks,
				Rollout:       c.Design.Rollout,
				Migration:     c.Design.Migration,
				OpenQuestions: c.Design.OpenQuestions,
			}
			// Only attach when at least one section is non-empty.
			if ds.Context != "" || ds.Goals != "" || ds.NonGoals != "" || ds.Decisions != "" ||
				ds.Risks != "" || ds.Rollout != "" || ds.Migration != "" || ds.OpenQuestions != "" {
				dc.Design = ds
			}
		}
		if c.Tasks != nil {
			for pi, p := range c.Tasks.Phases {
				ph := Phase{Number: p.Number, Name: p.Name, Items: []Item{}}
				for ii, it := range p.Items {
					it2 := Item{
						Text:       it.Text,
						Done:       it.Done,
						Kind:       string(it.Kind),
						Level:      pi,
						Key:        levelKey(pi, ii),
						Tags:       it.Tags,
						InlineRefs: it.InlineRefs,
					}
					if len(refs) > 0 {
						key := c.Name + "\x00" + p.Name + "\x00" + it.Text
						it2.ExternalRefs = refs[key]
					}
					ph.Items = append(ph.Items, it2)
				}
				dc.Phases = append(dc.Phases, ph)
			}
		}
		out = append(out, dc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return &Feed{Changes: out}
}

// JSON renders the feed as indented, deterministic JSON.
func (f *Feed) JSON() ([]byte, error) {
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
