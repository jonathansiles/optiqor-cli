# optiqor-cli — repo-local todo

This file tracks CLI-only work. The org-level roadmap that wires both
repos and the strategy docs is in [../todo.md](../todo.md); items
here are scoped to what lands inside this repo's `cmd/`, `internal/`,
or `pkg/`.

## Recently shipped

- [x] **2026-05-03 — Boxed cost-finding cards + signal bars** (playbook §4 Rule 1). `pkg/rules.Signal` carries quantitative evidence; renderer draws `█████░░░░` ratio bars inside per-finding cards. Graceful flat-layout fallback under 50 cols. Bug-fixed `visibleRuneCount` to iterate runes (was bytes) so multi-byte glyph alignment holds.
- [x] **2026-05-03 — `--roast` flag** (`internal/roast`). Static map of detector ID → snarky title for all 30 detectors. Hard rules preserved: no LLM, only `Title` mutated, accuracy disclosure exact.
- [x] **2026-05-03 — `score` letter grade + percentile** (`internal/analyze/grade.go`). Baked-in 100-sample calibration distribution; binary-search percentile lookup. `Grade` lands in JSON output too.
- [x] **2026-05-03 — `Category` first-class on Detector + Finding**. Drops the hardcoded `SecurityDetectorIDs` map; categorization is type-safe and audited in one place (`pkg/rules/categories.go`).
- [x] **2026-05-03 — Cost-first redesign of analyze output**. Branded header, executive summary, cost section sorted by USD savings descending, security section as a compact bonus block.
- [x] **2026-05-03 — Rebrand sevro → optiqor** across repo (151 files, including module path, package name, GitHub remote, tagline, README).

## Tier 1 — Launch anchors still open

- [ ] **Real `--share` upload endpoint** — CLI side already computes hash + posts to `sandbox.optiqor.dev/api/v1/share`; blocked on backend Phase 2 receiver. Don't over-build CLI side until the endpoint returns 2xx.
- [ ] **`compare` as side-by-side, not a `diff` alias** — playbook Feature 7 ("bitnami/postgresql vs cloudnative-pg") is press-bait. Needs a 2-column renderer + winner declaration.
- [ ] **Populate `Signal` on the remaining cost detectors** that have ratios but don't yet emit one: `oversized-cpu-limit`, `oversized-memory-limit`, `excessive-replica-count`, `tiny-cpu-request`, `tiny-memory-request`, `cpu-without-memory-request`, `memory-without-cpu-request`.

## Tier 2 — Distribution multipliers

- [ ] **`optiqor/actions` GitHub Action** wrapper (separate repo per playbook). Wraps `analyze --json`, posts a sticky PR comment.
- [ ] **Shell completions** — Cobra emits bash/zsh/fish for free; ship via goreleaser into the npm tarball.
- [ ] **Man page** — Cobra → `man optiqor` from the same registry.

## Tier 3 — Trust / enterprise gates

- [ ] **SBOM in releases** — `.goreleaser.yaml` `sbom:` stanza. SOC 2 / vendor-review gating.
- [ ] **cosign keyless OIDC provenance** — config exists but isn't wired to GitHub OIDC.
- [ ] **`--version --verbose`** — include commit, build date, Go version, target. Trivial Cobra wiring; matters for bug reports.
- [ ] **Per-detector golden fixtures** — `testdata/golden/` covers commands, not detectors. With 30 detectors and `pkg/rules` being public API the backend imports, drift will be silent without per-detector goldens.

## Hard rules — never violate

These are conditions for the OSS funnel to work. See [CLAUDE.md](CLAUDE.md) for the canonical list.

- **No LLM calls.** The CLI is a deterministic rule engine.
- **No telemetry by default.** Only `--share` egresses (opt-in).
- **Accuracy disclosure mandatory in every output.** Verbatim string; renderers must include it.
- **No proprietary backend code imported.** `go.mod` must never reference `github.com/optiqor/backend`.
- **`pkg/` is the stable public API.** Breaking changes go through semver and a deprecation notice.
