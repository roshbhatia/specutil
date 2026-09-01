# specutil

`specutil` reads local OpenSpec changes and projects them into review artifacts.
It performs no network I/O.

## Commands

| Command | Output |
|---|---|
| `specutil render` | RFC, design, or ticket Markdown |
| `specutil graph` | JSON, Mermaid, or DOT dependency graph |
| `specutil check` | Rubric findings for one change or the repository |
| `specutil next` | Runnable tasks after dependency checks |
| `specutil review` | Review decisions, comments, and drift |
| `specutil web` | A local HTML review page |

The repository must contain `openspec/changes/`. Optional settings live in
`openspec/specutil.yaml` and `openspec/config.yaml`.

## Install

```bash
nix run github:roshbhatia/specutil -- --help
```

Tagged releases publish `specutil` archives for Darwin and Linux.

## Examples

```bash
specutil render my-change --as rfc
specutil graph --as mermaid
specutil check my-change
specutil next my-change --as json
specutil review show my-change
specutil web --change my-change
```

## Development

```bash
nix develop
go test -race ./...
nix fmt -- --check
nix flake check
```

The flake exports the `discover-deps` and `review-change` skills for Sysinit.
