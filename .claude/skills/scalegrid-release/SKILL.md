---
name: scalegrid-release
description: How releases are cut for the terraform-provider-scalegrid repo (semantic-release + GoReleaser + GPG signing on push to main), how to merge so versions bump correctly, and how to regenerate docs. Use when merging a PR, cutting a release, debugging the release workflow, or running tfplugindocs.
---

# Releasing terraform-provider-scalegrid

Releases are fully automated: a push to `main` runs `.github/workflows/release.yml`, which uses
**semantic-release** to compute the next version from Conventional Commits and **GoReleaser** to
build, sign, and publish the GitHub release that the Terraform Registry ingests.

## Conventional Commits → version bump

semantic-release reads commit messages on `main`:

- `fix: ...` → **patch** (x.y.Z)
- `feat: ...` → **minor** (x.Y.0)
- `feat!: ...` or a `BREAKING CHANGE:` footer → **major** (X.0.0)
- `chore:`/`docs:`/`style:`/`test:`/`refactor:` → no release

## ⚠️ Merge with a MERGE COMMIT, not squash

This is the critical rule. semantic-release only sees commits that land on `main`. If you
**squash**, the PR's `feat:`/`fix:` subjects are replaced by the squash title and the bump can be
wrong or missing. **Always merge PRs with `merge_method: "merge"`** so the original Conventional
Commits reach `main`. (This is how every release in this repo has been cut.)

## What a successful release produces

GoReleaser publishes a GitHub release tagged `vX.Y.Z` with:

- 12 platform archives (`darwin/linux/windows/freebsd` × `386/amd64/arm/arm64`)
- `terraform-provider-scalegrid_<ver>_SHA256SUMS`
- `terraform-provider-scalegrid_<ver>_SHA256SUMS.sig` (GPG-signed; loopback pinentry)
- `terraform-provider-scalegrid_<ver>_manifest.json` (**must** be included in the checksums file
  for the Registry — this was a past fix)

semantic-release also pushes a `chore(release): X.Y.Z [skip ci]` commit/tag.

## Watching a release

After merging, poll the latest `release.yml` run on `main`:

```bash
rid=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
  ".../actions/workflows/release.yml/runs?branch=main&per_page=1" | jq '.workflow_runs[0].id')
# poll .../actions/runs/$rid until status == completed, conclusion == success
```

Then confirm the tag via the `releases/latest` API and that all assets above are present.

When merging multiple PRs, **merge them sequentially** — wait for each release run to finish
before merging the next, so the `[skip ci]` release commits don't race on `git push`.

## Docs (tfplugindocs)

Provider docs in `docs/` are generated from schema descriptions + `examples/`. Regenerate after
any schema change:

```bash
CHECKPOINT_DISABLE=1 tfplugindocs generate --provider-name scalegrid
```

- Requires a real `terraform` binary on PATH and `CHECKPOINT_DISABLE=1` (the checkpoint endpoint
  is blocked by the proxy). If `tfplugindocs` isn't installed, `go install
  github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest`.
- CI does **not** verify docs, but keep them in sync anyway. Commit the regenerated `docs/`.

## Pre-merge checklist

`gofmt -l internal/` (clean) · `go build ./...` · `go vet ./...` · `go test ./...` · docs regenerated.

## Repo hygiene

- Do **not** put model identifiers in commit messages, PR bodies, code, or any pushed artifact.
- Develop on a feature branch off the latest `main`; open a PR; never push straight to `main`.
- Don't create a PR or merge unless the user asks.
