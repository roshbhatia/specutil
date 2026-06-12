# tui-interaction Specification

## Purpose
TBD - created by archiving change interactive-visualizers. Update Purpose after archive.
## Requirements
### Requirement: Responsive layout
The TUI SHALL reflow its layout to the current terminal dimensions, recomputing
on every resize event, and SHALL scroll content that exceeds the available
height rather than truncating or overflowing it.

#### Scenario: Reflow on resize
- **WHEN** the terminal is resized while the TUI is running
- **THEN** the layout recomputes to the new width and height without visual
  overflow or clipping

#### Scenario: Scroll overflowing content
- **WHEN** a column or detail pane contains more rows than the visible height
- **THEN** the content is scrollable and no rows are silently dropped

#### Scenario: Narrow terminal degradation
- **WHEN** the terminal is too narrow to show all columns side by side
- **THEN** the layout degrades gracefully (e.g. to a single column or list)
  instead of breaking

### Requirement: Focus and context relationship highlighting
When a workstream node is selected, the TUI SHALL highlight that node's incoming
and outgoing dependency edges and immediate neighbors and de-emphasize unrelated
nodes, while preserving the layered-depth column layout. The TUI MUST NOT draw
free-form routed edges across the whole graph.

#### Scenario: Highlight on selection
- **WHEN** a node is selected in the graph view
- **THEN** its prerequisite and dependent neighbors are emphasized and unrelated
  nodes are dimmed

#### Scenario: Direction preserved by position
- **WHEN** the graph view renders
- **THEN** prerequisites appear in earlier depth columns and dependents in later
  columns

### Requirement: Terminal-adaptive palette
The TUI SHALL style itself with a palette that adapts to the terminal's background
(light or dark) rather than fixed color codes, so it is legible on both light and dark
terminals. Lifecycle, selection, neighbor-emphasis, progress, and done states SHALL each
have a defined light and dark variant.

#### Scenario: Adapts to terminal background
- **WHEN** the TUI runs in a light terminal and in a dark terminal
- **THEN** its colors adapt to each background and remain legible, without the user
  configuring a theme

### Requirement: Lifecycle-styled progress cards
Each workstream card on the board and in the graph columns SHALL convey, beyond its name,
the change's lifecycle (by color) and its completion progress (an inline progress bar with
a done/total count). Cards MUST remain readable when the terminal is too narrow for
side-by-side columns.

#### Scenario: Card shows lifecycle and progress
- **WHEN** a workstream card renders
- **THEN** it shows the change name, a lifecycle-derived color, and an inline progress bar
  with its done/total task count

### Requirement: Master-detail ticket panel
The TUI SHALL open a detail panel for a selected workstream showing its lifecycle,
progress, proposal why and what-changes, tasks grouped by phase with done indicators and
per-phase progress, task-kind markers (verify/apply/confirm) on items, and its depends-on
and blocks relationships presented beside the task checklist rather than only below it.
The panel SHALL open on an explicit action (Enter or click) and close on Esc, and its
layout MUST NOT break the panel border when composed beside the board.

#### Scenario: Open ticket detail
- **WHEN** the user presses Enter or clicks a workstream card
- **THEN** a detail panel opens showing that workstream's lifecycle, progress, why,
  what-changes, tasks-by-phase with per-phase progress and kind markers, and its
  depends-on/blocks shown alongside the checklist

#### Scenario: Close ticket detail
- **WHEN** the user presses Esc with the detail panel open
- **THEN** the panel closes and focus returns to the board

#### Scenario: Composed layout stays intact
- **WHEN** the detail panel is composed beside the board on a wide terminal
- **THEN** both the board and the panel render within the available width without broken
  borders or clipped content

