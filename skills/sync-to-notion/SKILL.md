---
name: sync-to-notion
description: Sync an OpenSpec change to Notion as an RFC or design doc using the deterministic specutil CLI for rendering/planning and the Notion MCP tools for the remote writes. Use when the user wants to push, sync, or publish an OpenSpec proposal as a Notion RFC or a spec as a Notion design doc.
license: MIT
compatibility: Requires the specutil binary on PATH and the Notion MCP server connected.
allowed-tools: Bash(specutil:*) mcp__claude_ai_Notion__notion-create-pages mcp__claude_ai_Notion__notion-update-page mcp__claude_ai_Notion__notion-fetch AskUserQuestion
metadata:
  author: specutil
  version: "1.0"
---

# Sync an OpenSpec change to Notion

This skill is the **apply** half of specutil's determinism boundary. The
`specutil` binary is pure and never touches the network: it renders the change
to Markdown deterministically and owns the local identity map (lockfile). You,
the agent, are the **only** thing that talks to Notion — via the Notion MCP
tools — and every local-state mutation routes back through `specutil lock set`.

Notion is **document-grained**, not task-grained: a change projects to one
Notion page (an RFC from the proposal, or a design doc from the design/specs),
not one page per task. So the unit of sync here is the whole rendered document.

## Markdown handling (resolved)

`notion-create-pages` and `notion-update-page` accept **Notion-flavored Markdown
directly** in their `content` field — there is no manual block conversion. Pass
the rendered Markdown straight through. For full fidelity on advanced
constructs (callouts, columns, toggles), first read the MCP resource
`notion://docs/enhanced-markdown-spec` via your client's resource interface and
adjust the rendered Markdown to match. Do **not** put the page title inside
`content`; set it under `properties.title`.

## Inputs

- `<change>`: the OpenSpec change name. If omitted and only one change exists,
  the binary auto-selects it; otherwise ask the user.
- Which projection: `rfc` (from the proposal) or `design` (from the design/specs).
- The destination Notion parent (a `page_id`, or a `data_source_id` if filing
  under a database). Ask the user once if unknown.
- `--auto` (optional): the user said "just go do it" — skip confirmation.
  Without it, **confirm before the remote write** (default).

## Flow: render → review → write → lock set

### 1. Render (deterministic, offline)

```bash
specutil render <change> --as rfc      # or: --as design
```

This emits clean Markdown to stdout. Capture it. The binary also computes a
content fingerprint you can record; use `specutil diff <change> --target notion`
to learn whether this is a create (no lock entry) or an update (content drifted
vs the lockfile).

### 2. Review (confirmation default)

Tell the user what you're about to publish: change name, projection (rfc/design),
destination parent, and create-vs-update. Unless `--auto` was given, use
AskUserQuestion to confirm before writing.

### 3. Write (the only network step)

- **Create** (no existing Notion page for this change+projection): call
  `mcp__claude_ai_Notion__notion-create-pages` with the chosen `parent`, the
  document title under `properties.title`, and the rendered Markdown as
  `content`. Capture the returned page ID/URL.
- **Update** (a lock entry already maps this change+projection to a page): call
  `mcp__claude_ai_Notion__notion-update-page` targeting that page ID, replacing
  its content with the freshly rendered Markdown.

Use a stable identity for the document, namespaced by projection — e.g.
`rfc` or `design` as the lock identity under the `notion` target (one page per
projection per change). Confirm before the write unless `--auto`.

### 4. Lock set (write-back through the binary)

After a successful create/update, record the page mapping:

```bash
specutil lock set <projection> <notion-page-id> \
  --change <change> --target notion \
  --content-hash <contentHash> --title "<doc title>"
```

where `<projection>` is `rfc` or `design`. Get `<contentHash>` from
`specutil diff <change> --target notion` (the entry's recorded hash) so the next
diff reports the doc as in-sync.

### 5. Verify

Re-run `specutil diff <change> --target notion`; a clean publish reports no
`new`/`changed` for that projection's identity.

## Guardrails

- The binary makes **no** network calls. Rendering is offline; only this skill
  writes to Notion.
- Page IDs live **only** in the lockfile, never in the OpenSpec source artifacts.
- Default to confirmation; `--auto` is the only thing that suppresses it.
- Pass rendered Markdown verbatim — do not hand-convert to blocks, and do not
  duplicate the title into `content`.
- One page per projection per change; re-syncing updates that page rather than
  creating duplicates.
