# cert-manager-webhook-ionos-cloud — agent notes

## CI/CD — decided, do not re-propose

Evaluated and rejected for this repo (2026-08). **Reopening one needs new data, not an argument.**

- **A container vulnerability scanner** (Trivy, grype, anchore). The image is
  `distroless/static-debian12` plus one statically-linked binary — there is no package surface to
  scan. Go-source vulnerabilities are covered by CodeQL and, later, `govulncheck`.
- **`ct lint` / `ct install` / `helm/chart-testing-action`.** The e2e job already installs the real
  chart into a real kind cluster alongside the real upstream cert-manager. Chart-testing would be a
  weaker duplicate.
- **SonarCloud, yamllint, markdownlint.** No Sonar project exists for this repo. CodeQL default setup
  covers the static-analysis need and is free here.
- **Moving chart publishing off GitHub Pages.** External users consume
  `https://ionos-cloud.github.io/cert-manager-webhook-ionos-cloud` as a Helm repo. Do not break it.
- **Repairing the key-based cosign path.** The published images are not signed today and the
  committed `cosign.pub` is a promise the pipeline never kept — go keyless (OIDC) rather than wiring
  up the old key.
- **Renaming workflow files or jobs for tidiness.** Every job name is also a branch-protection
  setting; a rename silently un-gates `main`.
- **Dropping the arm64 image or the two-arch manifest list.** External consumers use them.
- **Re-enabling the full 6-platform goreleaser snapshot on every PR.** 720s, 65% of a cold run, and
  there is no GOOS-specific code in this repo to justify it.
- **Docker layer caching.** Distroless base plus one `COPY` of a host-built binary. Nothing to cache.
- **`ionos-cloud/paas-github-actions` composites.** Not usable here: a public repo's workflow token
  cannot resolve `uses:` into a private repo. Public actions and local composites only.
- **Merge queue, test sharding, larger runners.** All evaluated. No working precedent to copy, and
  none addresses a measured problem in this repo.

### Wanted — do not read the above as "no new tooling"

- `timeout-minutes` and a `concurrency` group on all four workflows — neither exists today.
- An explicit least-privilege `permissions:` block per workflow. Three of the four have none, against
  a repo default of `write`.
- **CodeQL default setup, secret scanning, push protection** — settings toggles, no workflow file,
  free on a public repo, all three currently off.
- **Keyless cosign signing plus an SBOM on the release**, and a README fix: the documented
  `cosign verify` step cannot succeed against what is published today.
- **`zizmor`** alongside `actionlint`, non-blocking — it covers dangerous triggers, `uses:`
  SHA-pinning and over-broad `permissions:`, which `actionlint` does not.
- **`govulncheck` warning-only.** Never a required check here: fork PRs already report `skipped` on
  the sole required context, so a second source-level gate on top of that makes CI less trustworthy,
  not more.
- SHA-pinning for all 12 third-party action references across the two workflow files.

Tracking issue: ICNS-2158.
