## What changes

<!-- One sentence. -->

## Why

<!-- Link the issue, or explain the motivation if there's no issue. -->

## Notes for reviewers

<!-- Anything non-obvious: tricky edge cases, deliberate trade-offs, things to test. Delete if not needed. -->

---

- [ ] `go test ./...` passes
- [ ] Conventional commit title (`feat` / `fix` / `chore` / `docs` / `test`)
- [ ] If adding/changing a provider: unit tests cover the section-mapping contract
- [ ] If adding a sync target: a skill ships alongside it for remote writes
- [ ] If touching `plan` / `lock`: lock roundtrip test updated
