<!--
Thanks for contributing to optiqor-cli! Notes:

- PR title must follow Conventional Commits: `feat(scope): subject` (linted in CI).
- Every commit must be DCO-signed (`git commit -s`).
- Squash merges are the default — your PR title becomes the merge commit.
- CI runs build + race tests + lint on Ubuntu and macOS.
-->

## What

<!-- 1–2 sentences. What does this PR change? -->

## Why

Fixes #

## How

<!-- Notable design decisions. Keep it short. -->

## Testing

- [ ] `make lint test` passes locally
- [ ] `go test -race ./...` passes
- [ ] If a new rule: added a chart fixture in `testdata/` and a golden output
- [ ] Tested locally with: <!-- e.g. `./bin/optiqor analyze ./testdata/fixtures/basic-chart` -->

## Checklist

- [ ] PR title follows Conventional Commits (`feat(scope): subject`)
- [ ] All commits are DCO-signed (`git commit -s`)
- [ ] No unrelated changes pulled in
- [ ] Documentation updated where user-visible behavior changed
- [ ] No LLM calls or telemetry introduced
- [ ] No proprietary backend imports
