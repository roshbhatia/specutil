---
name: sync-to-linear
description: Sync an OpenSpec change's tasks to Linear issues using the deterministic specutil CLI for planning and the Linear MCP tools for the remote writes. Use when the user wants to push, sync, or reconcile OpenSpec tasks as Linear issues, or asks to "create Linear tickets" from a change.
license: MIT
compatibility: Requires the specutil binary on PATH and the Linear MCP server connected.
allowed-tools: Bash(specutil:*) mcp__claude_ai_Linear__save_issue mcp__claude_ai_Linear__get_issue mcp__claude_ai_Linear__list_teams AskUserQuestion
metadata:
  author: specutil
  version: "1.0"
---

# Sync OpenSpec tasks to Linear

This skill is the **apply** half of specutil's determinism boundary. The
`specutil` binary is pure and never touches the network: it parses the change,
computes a deterministic plan, and owns the local identity map (lockfile). You,
the agent, are the **only** thing that talks to Linear — via the Linear MCP
tools — and every local-state mutation routes back through `specutil lock set`.

Never reimplement planning, hashing, or diffing here. If you need to know what
to do, ask the binary.

## Inputs

- `<change>`: the OpenSpec change name (e.g. `add-auth`). If omitted and only one
  change exists, the binary auto-selects it; otherwise ask the user.
- `--auto` (optional): the user said "just go do it" — skip the per-operation
  confirmation prompt. Without it, **confirm before every remote write** (default).
- The target Linear `team` (required by Linear to create issues). Ask the user
  once if you don't already know it; reuse it for the whole run.

## Flow: plan → review → write → lock set

### 1. Plan (deterministic, offline)

```bash
specutil plan <change> --target linear
```

This emits `plan.json` to stdout:

```json
{
  "change": "add-auth",
  "target": "linear",
  "operations": [
    { "kind": "create", "identity": "01ac2154...", "contentHash": "b649...", "title": "Add login form", "ref": "1.2" },
    { "kind": "update", "identity": "06d998fb...", "externalId": "ENG-42", "contentHash": "eb73...", "title": "Harden session cookie", "ref": "1.5" },
    { "kind": "orphan", "identity": "ghost...", "externalId": "ENG-7", "title": "Removed task" }
  ]
}
```

Operation kinds:
- **create** — no lock entry; you will create a new Linear issue.
- **update** — lock entry exists but the task content changed; you will update
  the existing issue named by `externalId`.
- **orphan** — a lock entry whose task no longer exists locally; the binary will
  not delete anything. Surface it to the user and let them decide (close the
  Linear issue, leave it, or remove the lock entry).

After planning, always render the full ticket content:

```bash
specutil render <change> --as tickets
```

This emits Markdown for every task. Correlate each rendered block to its plan
operation by `ref` (e.g. `1.2`). Use the rendered Markdown as the Linear issue
`description` — it is the canonical body, not a nice-to-have. The bare task
title alone is never sufficient.

### 2. Review (confirmation default)

Summarize the plan to the user grouped by kind: N to create, M to update, K
orphaned. Unless `--auto` was given, use AskUserQuestion to confirm before
proceeding. List orphans explicitly — they are the only ones that imply remote
deletion intent, and the binary deliberately won't do that for you.

### 3. Write (the only network step)

For each operation, in plan order:

- **create**: call `mcp__claude_ai_Linear__save_issue` with `team`, `title`
  (from the op), and a Markdown `description` (the task text; pass literal
  newlines, not escape sequences). Do **not** pass `id`. Capture the returned
  issue identifier (e.g. `ENG-123`).
- **update**: call `save_issue` with `id: <externalId>` and the new
  `title`/`description`.
- **orphan**: only act if the user chose to. To close, call `save_issue` with
  `id: <externalId>` and `state` set to a done/canceled state.

Without `--auto`, confirm each write (or each batch) before issuing it.

### 4. Lock set (write-back through the binary)

Immediately after each successful create/update, record the mapping so the next
`plan`/`diff` sees it as in-sync:

```bash
specutil lock set <identity> <external-id> \
  --change <change> --target linear \
  --content-hash <contentHash> --title "<title>"
```

Use the `identity`, `contentHash`, and `title` straight from the plan op, and
the `external-id` you got back from Linear. For an orphan you closed, you may
leave the lock entry (history) or remove it per the user's choice; the binary
has no delete verb, so removing means editing the lockfile is out of scope —
prefer leaving it.

### 5. Verify

Re-run `specutil diff <change> --target linear`. A clean sync reports empty
`new` and `changed`. Remaining `orphaned` entries are expected if the user chose
to leave them.

## Guardrails

- The binary makes **no** network calls. If you find yourself wanting the binary
  to talk to Linear, stop — that work belongs here, in the skill.
- External IDs live **only** in the lockfile, never written back into the
  OpenSpec source artifacts.
- Default to confirmation; `--auto` is the only thing that suppresses it.
- Never invent identities or content hashes — always copy them from `plan.json`.
- Linear `save_issue` with no `id` creates; with `id` updates. Do not pass `id`
  on a create op or you will silently edit the wrong issue.
