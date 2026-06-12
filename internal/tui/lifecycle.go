// Package tui renders the workstream kanban and dependency-graph views with
// bubbletea + bubblezone. It imports the IR and graph packages directly and
// performs no network I/O — it only visualizes already-parsed local state.
package tui

import (
	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/lifecycle"
)

// Lifecycle and its states are re-exported from the shared lifecycle package so
// the TUI and the detail feed classify workstreams identically.
type Lifecycle = lifecycle.Lifecycle

const (
	Proposed = lifecycle.Proposed
	Active   = lifecycle.Active
	Archived = lifecycle.Archived
)

// LifecycleOrder is the left-to-right column order of the kanban board.
var LifecycleOrder = lifecycle.Order

// Progress counts completed and total task items across all phases.
func Progress(c *ir.Change) (done, total int) { return lifecycle.Progress(c) }

// Classify derives a change's lifecycle from its task progress.
func Classify(c *ir.Change) Lifecycle { return lifecycle.Classify(c) }
