# Contributing to specutil

Thanks for wanting to contribute. This is a short guide — specutil keeps things lean.

## Getting started

```bash
git clone https://github.com/roshbhatia/specutil
cd specutil
nix develop          # or: go 1.24+ if you prefer not to use Nix
go test ./...
```

## What's worth contributing

- **Bug fixes** — always welcome, especially with a regression test
- **New input providers** — there's a clean interface to implement; open a [New Provider](https://github.com/roshbhatia/specutil/issues/new?template=new_provider.yml) issue first
- **New sync targets** — same as providers; remote writes must live in a skill, not the binary
- **Docs / examples** — yes please

## Conventions

- **One concern per PR.** A bug fix doesn't need surrounding cleanup alongside it.
- **Conventional commits:** `feat`, `fix`, `chore`, `docs`, `test` — title-only, no body required.
- **Tests for providers:** every section-mapping path should have a unit test covering the tolerant-parse contract (absent sections emit warnings, never fail).
- **No network I/O in the binary.** The binary reads local files. Remote writes live in `skills/` and are driven by an agent. A new sync target means a new `skills/<name>/SKILL.md`.

## Adding an input provider

1. Create `internal/provider/<name>/<name>.go` implementing `provider.Provider`
2. Register it in `internal/registry/registry.go`
3. Add detection logic to `detect()` if the provider can be auto-detected
4. Ship tests in `internal/provider/<name>/<name>_test.go`

## Adding a sync target

1. Add the target name to `render/mapping.go`
2. Implement any plan-level fields in `syncplan/plan.go`
3. Write a skill in `skills/sync-to-<name>/SKILL.md` following the pattern of `sync-to-linear`

## Running checks

```bash
go test ./...          # unit tests
go vet ./...           # static analysis
nix flake check        # validate the flake (run before commits touching flake.nix)
```
