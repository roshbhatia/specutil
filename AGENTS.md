# Agent context for specutil

`specutil` is a local Go CLI for OpenSpec changes. The binary reads files and
emits Markdown, JSON, graphs, or HTML.

## Repository map

- `cmd/specutil`: executable entry point.
- `internal/cli`: Cobra commands.
- `internal/provider/openspec`: OpenSpec filesystem adapter.
- `internal/check`: rubric rules and schema presets.
- `internal/review`: decisions, annotations, and drift.
- `internal/render`, `internal/graph`, `internal/web`: output projections.
- `skills`: agent workflows that consume local output.

## Local boundary

Keep network access outside the binary. This keeps command output deterministic
and keeps credentials in the harness that owns them.

<example>
<bad>Add a Linear API client under `internal/cli`.</bad>
<good>Emit local JSON and let an agent skill call Linear.</good>
</example>

## Schema conventions

Put schema-specific extraction in `internal/extract` or rubric presets in
`internal/check`. This keeps all output projections schema-independent.

<example>
<bad>Branch on a schema name inside `internal/render`.</bad>
<good>Declare the marker in an extraction preset.</good>
</example>

## Identity

Use `internal/ident` for task and hunk identity. One implementation keeps review
comments attached after small text edits.

<example>
<bad>Hash task text again inside `internal/web`.</bad>
<good>Call `ident.Identity` and pass the result through the detail feed.</good>
</example>

## Checks

```bash
nix develop
go vet ./...
go test -race ./...
nix fmt -- --check
nix flake check
```
