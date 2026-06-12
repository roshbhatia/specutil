package syncplan

import (
	"sort"

	"github.com/roshbhatia/specutil/internal/ir"
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
				Identity:    Identity(p.Name, t.Text),
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

// Operation is a single create/update/orphan instruction. ExternalID is set for
// update and orphan (the existing remote object); Title/Ref describe the local
// source for create and update.
type Operation struct {
	Kind        OpKind `json:"kind"`
	Identity    string `json:"identity"`
	ExternalID  string `json:"externalId,omitempty"`
	ContentHash string `json:"contentHash,omitempty"`
	Title       string `json:"title,omitempty"`
	Ref         string `json:"ref,omitempty"`
}

// Plan is the deterministic, network-free projection of items against a lock.
type Plan struct {
	Change     string      `json:"change"`
	Target     string      `json:"target"`
	Operations []Operation `json:"operations"`
}

// BuildPlan diffs current items against the lock namespace for target and emits
// create/update/orphan operations. It performs no network I/O.
//
//   - identity absent from lock                -> create
//   - identity present, content hash differs   -> update (carries external ID)
//   - identity present, content hash unchanged  -> no operation (in sync)
//   - lock identity with no current item       -> orphan
func BuildPlan(change *ir.Change, lock *Lock, target string) Plan {
	items := TaskItems(change)
	current := make(map[string]bool, len(items))

	ops := make([]Operation, 0)
	for _, it := range items {
		current[it.Identity] = true
		ref, ok := lock.Get(target, it.Identity)
		switch {
		case !ok:
			ops = append(ops, Operation{
				Kind: OpCreate, Identity: it.Identity,
				ContentHash: it.ContentHash, Title: it.Title, Ref: it.Ref,
			})
		case ref.ContentHash != it.ContentHash:
			ops = append(ops, Operation{
				Kind: OpUpdate, Identity: it.Identity, ExternalID: ref.ExternalID,
				ContentHash: it.ContentHash, Title: it.Title, Ref: it.Ref,
			})
		}
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
	return Plan{Change: change.Name, Target: target, Operations: ops}
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
