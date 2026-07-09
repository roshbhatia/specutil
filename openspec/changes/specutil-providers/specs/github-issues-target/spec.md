## ADDED Requirements

### Requirement: plan --target github-issues emits GitHub-enriched operations
`specutil plan --target github-issues` SHALL emit the standard `Plan` JSON (change, target, operations array) with each `Operation` carrying an additional `github` object containing: `labels` (string array, phase-derived), `milestone` (string, change name), and `body` (string, pre-rendered issue body from the embedded template).

#### Scenario: GitHub plan emits enriched operations
- **WHEN** user runs `specutil plan --target github-issues --change my-feature`
- **THEN** stdout is valid JSON with `"target": "github-issues"` and each operation includes a `"github"` object with `labels`, `milestone`, and `body` fields

#### Scenario: Labels derived from phase name
- **WHEN** a task belongs to phase `"1. Foundation"`
- **THEN** `github.labels` includes `"phase:foundation"` (lowercased, spaces and punctuation stripped)

#### Scenario: Milestone set to change name
- **WHEN** the change is named `"my-feature"`
- **THEN** `github.milestone` equals `"my-feature"` for all operations in the plan

### Requirement: GitHub issue body pre-rendered by binary
The `github.body` field SHALL be rendered using the embedded `github-issues.md.tmpl` template, which SHALL be user-overridable via `--templates` (same override mechanism as rfc/design/tickets). The body SHALL include the task title, phase name, and change name. The binary renders the body deterministically; no network call is made.

#### Scenario: Body rendered with embedded template
- **WHEN** `--templates` is not set
- **THEN** `github.body` is rendered using the embedded `github-issues.md.tmpl`

#### Scenario: Body rendered with user override template
- **WHEN** `--templates ./my-templates/` and `./my-templates/github-issues.md.tmpl` exists
- **THEN** `github.body` is rendered using the user-provided template

#### Scenario: Plan output is human-readable
- **WHEN** user runs `specutil plan --target github-issues | jq '.[0].operations[0].github.body'`
- **THEN** the body is a readable markdown string suitable for inspection without calling the API

### Requirement: sync-to-github-issues skill orchestrates API calls
A `sync-to-github-issues` skill SHALL consume `specutil plan --target github-issues` JSON and apply each operation via `gh issue create` / `gh issue edit`, then call `specutil lock set` to record the mapping. The skill SHALL follow the confirm-then-apply pattern established by `sync-to-linear`.

#### Scenario: Create operation creates a GitHub issue
- **WHEN** the plan contains a `create` operation
- **THEN** the skill runs `gh issue create --title <title> --label <labels> --milestone <milestone> --body <body>` and calls `specutil lock set` with the returned issue number

#### Scenario: Update operation updates existing issue
- **WHEN** the plan contains an `update` operation with a stored `externalId` (issue number)
- **THEN** the skill runs `gh issue edit <number> --body <body>` and updates the lock

#### Scenario: Orphan operation flags but does not close
- **WHEN** the plan contains an `orphan` operation
- **THEN** the skill warns the user about the orphaned issue without closing it; auto-close requires `--auto` flag

### Requirement: github-issues target respects existing lock+diff contract
`specutil diff --target github-issues` SHALL work identically to `diff --target linear`: comparing current tasks against the lockfile and surfacing create/update/orphan operations. The `--target github-issues` namespace in the lockfile SHALL be independent of linear/notion namespaces.

#### Scenario: Diff shows drift for github-issues target
- **WHEN** a task has been renamed since the last sync to GitHub Issues
- **THEN** `specutil diff --target github-issues` shows the task as `update` with the stored issue number
