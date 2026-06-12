## ADDED Requirements

### Requirement: TUI workstream kanban
The `tui` verb SHALL render a bubbletea + bubblezone interface presenting changes as cards on a workstream-lifecycle kanban (e.g., proposed / active / archived), using bubblezone mouse zones for interaction. The TUI SHALL consume the IR in-process.

#### Scenario: Changes appear as lifecycle cards
- **WHEN** the user runs `specutil tui` in a repo with changes in different lifecycle states
- **THEN** each change appears as a card in the column matching its lifecycle state

#### Scenario: Empty repo renders without crashing
- **WHEN** `tui` is run in a repo with no changes
- **THEN** an empty board is shown with guidance rather than a crash

### Requirement: TUI layered dependency graph view
The TUI SHALL provide a dependency graph view that lays workstreams out in layers by dependency depth (not free-form edge routing).

#### Scenario: Layered ordering reflects dependencies
- **WHEN** change B depends on change A
- **THEN** B is placed in a deeper layer than A in the graph view

#### Scenario: Cyclic dependencies are surfaced
- **WHEN** the dependency manifest contains a cycle
- **THEN** the graph view indicates the cycle rather than looping indefinitely
