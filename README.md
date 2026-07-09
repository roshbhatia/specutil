# specutil

A deterministic CLI that reads spec changes from multiple input formats and projects them into renderable documents, sync plans, and visualizations. It performs no network I/O — all remote writes are delegated to an AI agent via shipped skills.

```
specutil [render|plan|diff|lock|graph|tui|web] [--from <provider>] [flags]
```

## What it does

```
Input providers (--from)           Core IR              Output
────────────────────────           ────────             ──────
openspec (default)   ──────────▶                ──▶ render  → RFC / design doc / tickets
bmad stories/*.md    ──────────▶  ir.Change     ──▶ plan    → create/update/orphan JSON
plan.md convention   ──────────▶                ──▶ diff    → lockfile delta
stdin / pipe         ──────────▶                ──▶ lock    → identity map
script adapters      ──────────▶                ──▶ graph   → DAG (json/mermaid/dot)
                                                ──▶ tui     → terminal kanban + graph
                                                ──▶ web     → HTML dashboard
```

The **determinism boundary** is the core invariant: everything predictable lives in the binary (pure functions, no network); everything stateful (remote writes, auth, drift reconciliation) is delegated to a shipped AI skill that drives the agent's Linear/Notion/GitHub MCP tools.

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

```bash
# OpenSpec project (auto-detected)
specutil render --as rfc
specutil plan --target linear

# BMAD project
specutil render --from bmad stories/story-1.1.md --as tickets
specutil plan --from bmad --target github-issues

# AI-generated plan.md
specutil render --from plan plan.md --as rfc

# Pipe from any tool
./my-adapter.sh my-change | specutil render --from stdin --as design

# Script adapter (declared in openspec/specutil.yaml)
specutil render --from jira --change PROJ-123 --as rfc
```

### Input provider auto-detection

When `--from` is omitted, specutil detects the provider from the repo layout:

| Signal | Provider |
|--------|----------|
| `openspec/changes/` directory | `openspec` |
| `stories/*.md` files | `bmad` |
| `plan.md` at root | `plan` |

### plan.md convention

Any markdown file that follows this structure works with `--from plan`:

```markdown
# change-name

## Why
One paragraph explaining the motivation.

## What Changes
- capability: description

## Tasks

### Phase 1: Foundation
- [ ] 1.1 First task
- [ ] 1.2 Second task
```

### Script adapters

Declare custom providers in `openspec/specutil.yaml`:

```yaml
providers:
  - name: jira
    command: "./hack/fetch-jira.sh {change}"
  - name: confluence
    command: "./hack/fetch-confluence.sh {change}"
```

Scripts receive the `--change` value as `{change}` and emit openspec-compatible markdown to stdout.

## Commands

### `render`

Projects a change's IR into a target artifact format.

```bash
specutil render [change] --as rfc|design|tickets [--from <provider>] [-o output.md]
```

| Flag | Description |
|------|-------------|
| `--as` | Target format: `rfc`, `design`, or `tickets` (required) |
| `--from` | Input provider: `openspec`, `bmad`, `plan`, `stdin`, or a script adapter name |
| `--change` | Change name (or pass as positional arg) |
| `--templates` | Override template directory |
| `-o` | Write to file instead of stdout |

```bash
# OpenSpec repo (auto-detected)
specutil render my-feature --as rfc -o docs/rfc.md

# BMAD story file
specutil render --from bmad stories/story-1.1.md --as tickets

# AI plan file
specutil render --from plan plan.md --as design

# Stdin (pipe from any tool)
cat plan.md | specutil render --from stdin --as rfc
```

### `plan`

Emits a deterministic create/update/orphan plan for syncing to a remote target.

```bash
specutil plan [change] --target linear|notion|github-issues [--from <provider>] [-o plan.json]
```

The plan is a list of operations (`create`, `update`, `orphan`) keyed by stable content hashes, with no network calls made by the binary itself.

For `--target github-issues`, each operation includes a pre-rendered `github.body` field (markdown), `github.labels` derived from the phase name, and `github.milestone` set to the change name. The `sync-to-github-issues` skill reads these fields directly — no re-templating required.

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
- Unified overview with lifecycle board columns — click a card to inline-expand its full detail; Graph view for the full DAG workbench with node inspector
- Data inlined; Chart.js loaded from a version-pinned, SRI-protected CDN at view time
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
cmd/specutil/             CLI entrypoint
internal/
  ir/                     Intermediate representation (framework-agnostic types)
  provider/               Provider port definition
    openspec/             OpenSpec adapter (discovery + loading)
    bmad/                 BMAD story file adapter
    plan/                 plan.md convention adapter (+ stdin)
    script/               User-defined script adapter
  registry/               Provider selection and auto-detection
  parse/                  goldmark-based lenient markdown parser
  render/                 Artifact rendering (mapping + templates + Sprig)
  graph/                  Cross-change DAG (build, project, suggest)
  detail/                 Per-change detail feed for visualizers
  syncplan/               Plan/diff/lock (content-hash identity)
  tui/                    Bubbletea terminal UI
  web/                    Static HTML generator
  lifecycle/              Lifecycle classification helpers
  cli/                    Cobra command wiring
skills/
  sync-to-linear/         Linear sync skill
  sync-to-notion/         Notion sync skill
  sync-to-github-issues/  GitHub Issues sync skill
openspec/                 specutil's own OpenSpec change (specutil-providers)
```

## License

MIT
