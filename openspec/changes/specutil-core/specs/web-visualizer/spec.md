## ADDED Requirements

### Requirement: Static web dependency graph
The `serve` verb SHALL produce a lightweight, static web site that renders the cross-change dependency DAG, consuming the `graph.json` feed and rendering it via Mermaid in v1. The site SHALL require no backend and SHALL function offline.

#### Scenario: Static site renders the DAG
- **WHEN** the user runs `specutil serve` and opens the site
- **THEN** the workstream dependency graph is displayed using Mermaid, served from static assets only

#### Scenario: Works offline
- **WHEN** the site is opened with no network access
- **THEN** the graph still renders from the locally-served assets and `graph.json`

#### Scenario: Empty graph renders gracefully
- **WHEN** there are no dependency edges
- **THEN** the site shows the changes (or an empty-state message) without erroring

### Requirement: Renderer is swappable
The web layer SHALL consume the canonical `graph.json` such that the Mermaid renderer can later be replaced (e.g., with Cytoscape.js) without changing the data feed.

#### Scenario: Data feed independent of renderer
- **WHEN** the rendering library is changed
- **THEN** the same `graph.json` feed is consumed without modification to its schema
