---
name: sync-to-github-issues
description: Sync an OpenSpec change's tasks to GitHub Issues using specutil CLI for planning and the gh CLI for the remote writes. Use when the user wants to push, sync, or reconcile OpenSpec tasks as GitHub Issues, or asks to "create GitHub issues" from a change.
license: MIT
compatibility: Requires the specutil binary on PATH and the gh CLI authenticated.
allowed-tools: Bash(specutil:*) Bash(gh:*) AskUserQuestion
metadata:
  author: specutil
  version: "1.0"
---

# Sync OpenSpec tasks to GitHub Issues

The `specutil` binary reads your local change artifacts and produces a plan
with pre-rendered issue bodies; it never touches the network. You, the agent,
are the **only** thing that calls the GitHub API — via `gh issue` — and every
local-state mutation routes back through `specutil lock set`.

Never reimplement planning, label derivation, or body rendering. If you need
to know what to do, ask the binary.

## Inputs

- `<change>`: the change name (e.g. `add-auth`). If omitted and only one change
  exists, the binary auto-selects it; otherwise ask the user.
- `--auto` (optional): skip the per-operation confirmation prompt.
- `--repo` (optional): GitHub repo in `owner/name` form (e.g. `acme/backend`).
  If omitted, `gh` uses the current repo.

## Flow: plan → review → write → lock set

### 1. Plan

```bash
specutil plan --change <change> --target github-issues
```

This emits JSON to stdout. Each operation in the `operations` array includes a
`github` object with pre-rendered fields:

```json
{
  "change": "add-auth",
  "target": "github-issues",
  "operations": [
    {
      "kind": "create",
      "identity": "01ac2154...",
      "contentHash": "b649...",
      "title": "1.1 Add login form",
      "ref": "1.1",
      "github": {
        "labels": ["phase:foundation"],
        "milestone": "add-auth",
        "body": "**Change:** add-auth\n**Phase:** Foundation\n..."
      }
    },
    {
      "kind": "update",
      "identity": "06d998fb...",
      "externalId": "42",
      "contentHash": "eb73...",
      "title": "1.2 Harden session cookie",
      "ref": "1.2",
      "github": {
        "labels": ["phase:foundation"],
        "milestone": "add-auth",
        "body": "..."
      }
    },
    {
      "kind": "orphan",
      "identity": "ghost...",
      "externalId": "7"
    }
  ]
}
```

Operation kinds:
- **create** — no lock entry; you will create a new GitHub Issue.
- **update** — lock entry exists but task content changed; you will update the
  existing issue numbered by `externalId`.
- **orphan** — a lock entry whose task no longer exists locally. The binary
  will not close anything. Surface it to the user and let them decide.

### 2. Review (confirmation default)

Summarize the plan to the user grouped by kind: N to create, M to update, K
orphaned. Unless `--auto` was given, use AskUserQuestion to confirm before
proceeding. List orphans explicitly — they imply potential remote deletion and
must be user-approved.

### 3. Write (the only network step)

For each operation, in plan order:

**create:**
```bash
gh issue create \
  --title "<op.title>" \
  --body "<op.github.body>" \
  --label "<label1>" --label "<label2>" \
  --milestone "<op.github.milestone>"
```
Use `jq` to read fields from the plan JSON. `gh issue create` prints the new
issue URL; extract the issue number from it (last path segment).

**update:**
```bash
gh issue edit <op.externalId> \
  --body "<op.github.body>"
```
Only update the body (the title is set at creation and typically stays stable).
If the title changed, add `--title "<op.title>"`.

**orphan:**
Warn the user with the issue URL (`gh issue view <op.externalId> --json url`).
Do **not** close the issue unless `--auto` is set:
```bash
gh issue close <op.externalId>   # only with --auto
```

Without `--auto`, confirm each write before issuing it.

### 4. Lock set (write-back through the binary)

Immediately after each successful create/update, record the mapping:

```bash
specutil lock set <identity> <issue-number> \
  --change <change> --target github-issues \
  --content-hash <contentHash> --title "<title>"
```

Use `identity`, `contentHash`, and `title` from the plan op, and the issue
number from the `gh` response. Never invent these values.

For orphans you closed (with `--auto`), you may leave the lock entry in place —
the binary has no delete verb.

### 5. Verify

```bash
specutil diff --change <change> --target github-issues
```

A clean sync reports empty `new` and `changed`. Remaining `orphaned` entries
are expected if the user chose to leave them open.

## Milestone setup

GitHub milestones must exist before `gh issue create --milestone` can reference
them. If the milestone `<change>` does not yet exist, create it first:

```bash
gh api repos/{owner}/{repo}/milestones \
  --method POST \
  --field title="<change>"
```

Check existence before creating to avoid duplicates.

## Label setup

GitHub labels must exist before they can be applied. Check for missing labels
and create them if needed:

```bash
gh label list | grep "phase:"
gh label create "phase:foundation" --color "0075ca"
```

Create all required phase labels before the first `gh issue create`.

## Guardrails

- The binary makes **no** network calls. If you find yourself wanting the
  binary to talk to GitHub, stop — that work belongs here, in the skill.
- External IDs (issue numbers) live **only** in the lockfile, never written
  back into the change source artifacts.
- Default to confirmation; `--auto` is the only thing that suppresses it.
- Never invent identities or content hashes — always copy them from the plan.
- `gh issue create` with no `--issue-number` creates; `gh issue edit <number>`
  updates. Do not confuse the two.
- Use the `github.body` from the plan JSON verbatim — it is pre-rendered by
  the binary. Do not re-template it.
