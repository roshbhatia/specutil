package syncplan

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/render"
)

// Item is a plannable unit derived from a change. For ticketing targets the
// units are tasks; the abstraction leaves room for document-level items later.
type Item struct {
	Identity    string
	ContentHash string
	Title       string
	Ref         string // human-facing source locator, e.g. the task number
}

// TaskItems projects a change's tasks into plannable items. Identity is built
// from the phase name and task text (renumber-stable); ContentHash fingerprints
// the exact text for drift detection.
func TaskItems(change *ir.Change) []Item {
	if change == nil || change.Tasks == nil {
		return nil
	}
	var items []Item
	for _, p := range change.Tasks.Phases {
		for _, t := range p.Items {
			items = append(items, Item{
				// Include Phase.Number so number-only headings ("## 1." and "## 2.") and
				// same-titled phases with different numbers don't collapse to the same key.
				Identity:    Identity(p.Number+" "+p.Name, t.Text),
				ContentHash: ContentHash(t.Text),
				Title:       t.Text,
				Ref:         t.ID,
			})
		}
	}
	return items
}

// OpKind is a planned operation against the target system.
type OpKind string

const (
	OpCreate OpKind = "create"
	OpUpdate OpKind = "update"
	OpOrphan OpKind = "orphan"
)

// GitHubFields carries pre-rendered GitHub-specific metadata for the
// github-issues plan target. Populated only when target == "github-issues".
type GitHubFields struct {
	Labels    []string `json:"labels"`
	Milestone string   `json:"milestone"`
	Body      string   `json:"body"`
}

// Operation is a single create/update/orphan instruction. ExternalID is set for
// update and orphan (the existing remote object); Title/Ref describe the local
// source for create and update.
type Operation struct {
	Kind        OpKind        `json:"kind"`
	Identity    string        `json:"identity"`
	ExternalID  string        `json:"externalId,omitempty"`
	ContentHash string        `json:"contentHash,omitempty"`
	Title       string        `json:"title,omitempty"`
	Ref         string        `json:"ref,omitempty"`
	GitHub      *GitHubFields `json:"github,omitempty"`
}

// BuildPlanOptions carries optional configuration for BuildPlan.
type BuildPlanOptions struct {
	// TemplateOverrideDir is passed to the render engine for github-issues body
	// rendering. Empty means use the embedded default.
	TemplateOverrideDir string
}

// Plan is the deterministic, network-free projection of items against a lock.
type Plan struct {
	Change     string      `json:"change"`
	Target     string      `json:"target"`
	Operations []Operation `json:"operations"`
	Warnings   []string    `json:"warnings,omitempty"`
}

// BuildPlan diffs current items against the lock namespace for target and emits
// create/update/orphan operations. It performs no network I/O.
//
//   - identity absent from lock                -> create
//   - identity present, content hash differs   -> update (carries external ID)
//   - identity present, content hash unchanged  -> no operation (in sync)
//   - lock identity with no current item       -> orphan
func BuildPlan(change *ir.Change, lock *Lock, target string) Plan {
	plan, _ := BuildPlanWithOptions(change, lock, target, BuildPlanOptions{})
	return plan
}

// BuildPlanWithOptions is the full form of BuildPlan with optional configuration.
// It returns an error only when the github-issues body template fails to render;
// partial success is not possible — either all bodies render or none do.
func BuildPlanWithOptions(change *ir.Change, lock *Lock, target string, opts BuildPlanOptions) (Plan, error) {
	items := TaskItems(change)
	current := make(map[string]bool, len(items))

	// Build a phase-lookup map for github-issues label derivation.
	phaseByRef := buildPhaseByRef(change)

	ops := make([]Operation, 0)
	var planWarnings []string
	for _, it := range items {
		current[it.Identity] = true
		ref, ok := lock.Get(target, it.Identity)
		var op Operation
		switch {
		case !ok:
			op = Operation{
				Kind: OpCreate, Identity: it.Identity,
				ContentHash: it.ContentHash, Title: it.Title, Ref: it.Ref,
			}
		case ref.ContentHash != it.ContentHash:
			op = Operation{
				Kind: OpUpdate, Identity: it.Identity, ExternalID: ref.ExternalID,
				ContentHash: it.ContentHash, Title: it.Title, Ref: it.Ref,
			}
		default:
			continue
		}
		if target == "github-issues" {
			gh, warn, err := buildGitHubFields(change, it, phaseByRef[it.Ref], opts.TemplateOverrideDir)
			if err != nil {
				return Plan{}, err
			}
			if warn != nil {
				planWarnings = append(planWarnings, warn.Msg)
			}
			op.GitHub = gh
		}
		ops = append(ops, op)
	}

	for _, id := range lock.Identities(target) {
		if !current[id] {
			ref, _ := lock.Get(target, id)
			ops = append(ops, Operation{
				Kind: OpOrphan, Identity: id, ExternalID: ref.ExternalID,
			})
		}
	}

	sortOps(ops)
	return Plan{Change: change.Name, Target: target, Operations: ops, Warnings: planWarnings}, nil
}

// buildPhaseByRef builds a map from task ref (e.g. "1.2") to phase name for
// github-issues label derivation.
func buildPhaseByRef(change *ir.Change) map[string]string {
	m := map[string]string{}
	if change.Tasks == nil {
		return m
	}
	for _, p := range change.Tasks.Phases {
		for _, it := range p.Items {
			m[it.ID] = p.Name
		}
	}
	return m
}

// buildGitHubFields populates the github-specific fields for an operation.
// The warning return is non-nil when an override template was requested but not
// found and the embedded default was used instead — callers may emit it.
func buildGitHubFields(change *ir.Change, it Item, phaseName, overrideDir string) (*GitHubFields, *ir.Warning, error) {
	body, warn, err := render.RenderIssueBody(change, render.IssueBodyData{
		PhaseName: phaseName,
		TaskRef:   it.Ref,
		TaskTitle: it.Title,
	}, overrideDir)
	if err != nil {
		return nil, nil, fmt.Errorf("github-issues body for %q: %w", it.Ref, err)
	}

	return &GitHubFields{
		Labels:    deriveGitHubLabels(phaseName),
		Milestone: change.Name,
		Body:      body,
	}, warn, nil
}

// labelCleanRe strips non-alphanumeric characters for label normalization.
var labelCleanRe = regexp.MustCompile(`[^a-z0-9]+`)

// deriveGitHubLabels converts a phase name to a GitHub label slice.
// "1. Foundation" → ["phase:foundation"]
func deriveGitHubLabels(phaseName string) []string {
	if phaseName == "" {
		return nil
	}
	// Strip leading "N." numbering.
	name := regexp.MustCompile(`^\d+\.?\s*`).ReplaceAllString(phaseName, "")
	name = strings.ToLower(name)
	name = labelCleanRe.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return nil
	}
	return []string{"phase:" + name}
}

// sortOps orders operations deterministically: by kind (create, update,
// orphan), then identity.
func sortOps(ops []Operation) {
	rank := map[OpKind]int{OpCreate: 0, OpUpdate: 1, OpOrphan: 2}
	sort.Slice(ops, func(i, j int) bool {
		if rank[ops[i].Kind] != rank[ops[j].Kind] {
			return rank[ops[i].Kind] < rank[ops[j].Kind]
		}
		return ops[i].Identity < ops[j].Identity
	})
}
