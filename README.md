# specutil

A deterministic CLI that projects [OpenSpec](https://openspec.dev) change artifacts into renderable documents, sync plans, and visualizations. It performs no network I/O — all remote writes are delegated to an AI agent via shipped skills.

```
specutil [render|plan|diff|lock|graph|tui|serve] [flags]
```

## What it does

```
OpenSpec changes (openspec/changes/<name>/)
       │
       ▼
  specutil (pure Go binary, zero network I/O)
       │
       ├── render   → RFC / design doc / ticket markdown
       ├── plan     → create/update/orphan plan.json
       ├── diff     → local IR vs lockfile delta
       ├── lock     → identity map (content hash → external ID)
       ├── graph    → cross-change DAG (json / mermaid / dot)
       ├── tui      → terminal kanban + dependency graph
       └── serve    → static HTML dashboard (inline SVG DAG)
```

The **determinism boundary** is the core invariant: everything predictable lives in the binary (pure functions, no network); everything stateful (remote writes, auth, drift reconciliation) is delegated to a shipped AI skill that drives the agent's Linear/Notion MCP tools.

## Installation

### Nix (recommended)

```bash
# Run without installing
nix run github:roshbhatia/specutil -- --help

# Install via nix profile
nix profile install github:roshbhatia/specutil
```

### From source

```bash
git clone https://github.com/roshbhatia/specutil
cd specutil
nix develop          # enter dev shell with Go toolchain
go build -o specutil ./cmd/specutil
```

### Via hack/install.sh

```bash
./hack/install.sh    # builds and installs to $GOPATH/bin
```

## Quickstart

Given an OpenSpec repo with at least one change:

```
my-repo/
└── openspec/
    └── changes/
        └── my-feature/
            ├── proposal.md
            ├── tasks.md
            └── design.md
```

```bash
# Open the interactive dashboard in your browser
specutil serve

# Open the terminal kanban
specutil tui

# Render as an RFC
specutil render --as rfc

# Emit a sync plan for Linear
specutil plan --target linear
```

## Commands

### `render`

Projects a change's IR into a target artifact format.

```bash
specutil render [change] --as rfc|design|tickets [-o output.md]
```

| Flag | Description |
|------|-------------|
| `--as` | Target format: `rfc`, `design`, or `tickets` (required) |
| `--change` | Change name (or pass as positional arg) |
| `--templates` | Override template directory |
| `-o` | Write to file instead of stdout |

```bash
# Render the change named "my-feature" as an RFC
specutil render my-feature --as rfc -o docs/rfc.md

# Single-change repo: no name needed
specutil render --as design
```

### `plan`

Emits a deterministic create/update/orphan plan for syncing to a remote target.

```bash
specutil plan [change] --target linear|notion [-o plan.json]
```

The plan is a list of operations (`create`, `update`, `orphan`) keyed by stable content hashes, with no network calls made by the binary itself.

### `diff`

Compares the local IR against the per-change lockfile to show what has drifted.

```bash
specutil diff [change] --target linear
```

### `lock`

Manages the identity map between content hashes and external IDs (e.g., Linear issue IDs).

```bash
# Read an entry (exits 3 if not found)
specutil lock get <identity> --target linear --change my-feature

# Write an entry
specutil lock set <identity> <external-id> --target linear --change my-feature \
  --content-hash <hash> --title "Issue title"
```

The lockfile lives at `openspec/changes/<name>/specutil.lock.yaml` and is never written to source — only managed by `lock set`.

### `graph`

Projects the cross-change dependency DAG.

```bash
specutil graph --as json|mermaid|dot [-o graph.json]
specutil graph --suggest    # infer candidate edges without mutating the manifest
```

Dependencies come from `openspec/specutil.yaml`. Use `--suggest` to get inferred candidates from shared capabilities (does not write the file).

### `tui`

Opens the terminal kanban and dependency graph.

```bash
specutil tui
```

- Left panel: lifecycle kanban (proposed / active / archived)
- Right panel: layered-by-depth dependency graph
- Mouse zones via bubblezone; keyboard navigation supported

### `serve`

Generates and opens a static HTML dashboard.

```bash
specutil serve [-o output.html] [--open=false]
```

The page is a single self-contained HTML file:
- Cross-change dependency DAG as binary-rendered inline SVG
- Per-change progress, remaining tasks, per-phase chart
- Overview / Board / Graph views
- Data inlined; Pico CSS + Chart.js loaded from a version-pinned, SRI-protected CDN at view time
- Binary performs zero network I/O

## OpenSpec Integration

specutil reads from the standard OpenSpec directory layout. No changes to your OpenSpec authoring workflow are required.

### Directory layout

```
openspec/
├── config.yaml              # OpenSpec project config (unchanged)
├── specutil.yaml            # Cross-change dependency manifest (new)
└── changes/
    └── my-feature/
        ├── proposal.md
        ├── tasks.md
        ├── design.md
        ├── specs/
        │   └── my-cap/
        │       └── spec.md
        └── specutil.lock.yaml   # Written by `lock set` (gitignore or commit)
```

### proposal.md structure

specutil reads `## Why`, `## What Changes`, `## Capabilities`, and `## Impact` sections.

```markdown
## Why

We need X because Y.

## What Changes

- Add `foo` command that does X
- Modify `bar` to support Y

## Capabilities

### New Capabilities
- `my-capability`: Does the thing.

### Modified Capabilities
- `existing-cap`: Extends to support Y.

## Impact

- **New code:** `internal/foo`
- **Dependencies:** none new
```

### tasks.md structure

Phases are `## N. Phase Name` headings; tasks are `- [x]`/`- [ ]` checkboxes. Tasks can be tagged with `verify:`, `apply:`, or `confirm:` prefixes.

```markdown
## 1. Foundation

- [x] 1.1 Initialize the module
- [x] 1.2 verify: Build succeeds in dev shell
- [ ] 1.3 Add the core command

## 2. Integration

- [ ] 2.1 apply: Push branch and open PR (impactful)
- [ ] 2.2 confirm: CI green and PR reflects intended diff
```

### specutil.yaml (cross-change dependencies)

```yaml
# openspec/specutil.yaml
edges:
  - from: auth-redesign      # prerequisite
    to: user-profile-update  # depends on auth-redesign
  - from: auth-redesign
    to: session-management
```

Use `specutil graph --suggest` to see inferred candidates from shared capabilities. Edges must be confirmed and written manually (or via the agent skill) — `--suggest` never mutates the file.

### Lockfile

`specutil.lock.yaml` maps stable content hashes to external IDs. It is written only by `lock set` and read by `plan`/`diff`. Commit it alongside your changes or add it to `.gitignore` depending on your workflow.

```yaml
# openspec/changes/my-feature/specutil.lock.yaml
linear:
  abc123def:
    external_id: LIN-456
    content_hash: abc123def
    title: "Initialize the module"
```

## Integration Skills

Two skills ship in-repo to orchestrate specutil with Linear and Notion. Install them by symlinking into your skills directory or via home-manager.

### sync-to-linear

```
skills/sync-to-linear/SKILL.md
```

Flow: `plan` → review → MCP write → `lock set`. Each write is individually confirmed unless `--auto` is passed.

### sync-to-notion

```
skills/sync-to-notion/SKILL.md
```

Same flow as `sync-to-linear`, adapted for Notion pages and blocks.

## Examples

### End-to-end: render → plan → sync

```bash
# 1. Check what's in the change
specutil render --as rfc

# 2. See what would be created in Linear
specutil plan --target linear

# 3. Diff against what's already synced
specutil diff --target linear

# 4. Invoke the sync skill (agent-driven; prompts for confirmation)
# (from Claude Code or another AI shell)
# > run sync-to-linear
```

### Multi-change repository

```bash
# See all changes and their dependencies
specutil graph --as mermaid

# Infer missing dependency edges
specutil graph --suggest

# Open the visual dashboard
specutil serve
```

### Custom templates

```bash
# Override the RFC template
mkdir my-templates
cp $(go env GOPATH)/... my-templates/rfc.md.tmpl  # edit as needed
specutil render --as rfc --templates my-templates/
```

## Development

```bash
# Enter the dev shell
nix develop

# Build
go build ./...

# Test
go test ./...

# Format Nix files
nix fmt

# Format shell scripts
task fmt:sh

# Validate the flake
nix flake check
```

### Package layout

```
cmd/specutil/           CLI entrypoint
internal/
  ir/                   Intermediate representation (framework-agnostic types)
  provider/openspec/    OpenSpec adapter (discovery + loading)
  parse/                goldmark-based lenient markdown parser
  render/               Artifact rendering (mapping + templates)
  graph/                Cross-change DAG (build, project, suggest)
  detail/               Per-change detail feed for visualizers
  syncplan/             Plan/diff/lock (content-hash identity)
  tui/                  Bubbletea terminal UI
  web/                  Static HTML generator
  lifecycle/            Lifecycle classification helpers
  cli/                  Cobra command wiring
skills/
  sync-to-linear/       Linear sync skill
  sync-to-notion/       Notion sync skill
openspec/               specutil's own OpenSpec change (specutil-core)
```

## License

MIT
