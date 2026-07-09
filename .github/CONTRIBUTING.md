# Contributing to specutil

## Philosophy

specutil is intentionally small. The binary is a pure, local projection tool —
no network calls, no auth, no sync logic. Remote writes live in AI skills, not
in the binary. Contributions that respect this boundary are welcome;
contributions that blur it are not.

I maintain this project in my spare time. I review PRs carefully, and I will
close ones that don't fit the scope or that add complexity I'm not prepared to
maintain.

## Before you start

**If you are an AI agent:** Read [`AGENTS.md`](../AGENTS.md) first. It has
architecture context, testing patterns, and an explicit list of what not to do.
This file is for humans; `AGENTS.md` is for machines.

**If you are a human:** Open an issue before writing code for anything beyond a
trivial bug fix. A quick conversation about scope saves everyone time.

## What I'll accept

- **Bug fixes** — always, especially with a regression test.
- **New input providers** — implement `provider.Provider`, register it, ship
  tests. Open a [New Provider](https://github.com/roshbhatia/specutil/issues/new?template=new_provider.yml)
  issue first so we align on the interface before you build.
- **New sync targets** — a `skills/sync-to-<name>/SKILL.md`, not a new verb.
  The binary emits plan JSON; the skill drives the MCP tools. Open an issue first.
- **Docs and examples** — yes, no issue needed.

## What I won't accept

- A `sync` verb or any network I/O in the binary. This is the one hard rule.
- New verbs without a clear, scoped use case and prior discussion.
- Refactors that touch many files without a concrete correctness or
  maintainability reason.
- Dependencies added without discussion.

## Getting started

```bash
git clone https://github.com/roshbhatia/specutil
cd specutil
nix develop          # preferred: Go toolchain + all tools in one command
# or: go 1.24+ works if you don't use Nix
go test -race ./...
```

## Conventions

- **Conventional commits, title-only.** `feat`, `fix`, `chore`, `docs`, `test`.
  No body required.
- **One concern per PR.** Bug fix ≠ cleanup; a formatting change alone is not a
  PR.
- **No comments unless the WHY is non-obvious.** Never restate what the code
  already says.
- **Tests for providers.** Every section-mapping path needs a unit test covering
  the tolerant-parse contract: absent sections emit warnings, never errors.
- **No network I/O in the binary.** If you need to touch an API, that belongs
  in a skill.

## Adding an input provider

1. Create `internal/provider/<name>/<name>.go` implementing `provider.Provider`
2. Register in `internal/registry/registry.go`
3. Add auto-detection logic to `detect()` if applicable
4. Ship tests in `internal/provider/<name>/<name>_test.go`

## Adding a sync target

1. Add the target constant to `syncplan/plan.go`
2. Write `skills/sync-to-<name>/SKILL.md` following the pattern of
   `skills/sync-to-linear/SKILL.md`
3. No binary changes needed for the actual write — that lives in the skill

## Running checks

```bash
go test -race ./...    # tests with race detector
go vet ./...           # static analysis
nix flake check        # validate the flake (run before touching flake.nix)
```
