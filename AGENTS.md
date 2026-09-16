# CI/CD decisions

This is a **public** repository. Its CI/CD design has been through a security
review (findings CMW-01, CMW-02, CMW-03) and the decisions below were measured,
not guessed. Do not silently redo, revert, or "improve" them.

## Do not reconsider

Each item below was measured and rejected. Reopening requires a new
measurement, not an argument — put it on ICNS-2158 and update
"CI/CD decisions", which is the authority.

- Do not use any `paas-github-actions` asset or Vault from this repo's
  workflows — structurally impossible from a public repo (a public repo's
  workflow token cannot resolve `uses:` into a private repo, and Vault access
  is scoped to trusted/internal runners only).
- Do not add a container vulnerability scanner.
- Do not add `ct lint` / `ct install`.
- Do not add SonarCloud, yamllint, or markdownlint.
- Do not move chart publishing to Harbor.
- Do not repair the key-based cosign signing path (`COSIGN_PRIVATE_KEY` /
  `COSIGN_KEY_PASSWORD`). Images are signed keylessly instead.
- Do not rename workflows or jobs for tidiness.
- Do not drop the arm64 image or the two-arch manifest.
- Do not re-enable the 6-platform (3 OS x 2 arch) snapshot per PR — the PR
  snapshot only needs to prove the linux/amd64 + linux/arm64 docker path
  builds; full cross-OS archives are exercised at tag-release time.
- `govulncheck` is warning-only and must never become a required check here —
  the merge gate is already fragile (fork PRs legitimately report jobs as
  skipped pending maintainer approval), and a second source-level gate on top
  of that makes CI less trusted, not more.

## Sequencing constraint for environment/branch-protection changes

`ci-tests` environment has required reviewers. `main` branch protection has
`enforce_admins: true`. If the sole required status check is ever a job that
sits behind that environment's reviewer gate, every PR and every push to
main (including dependabot) blocks on a human approval that may never come.

The safe order is:
1. Split jobs along the secret boundary (secret-free jobs run unconditionally
   on `pull_request`; credential-bearing jobs stay on `pull_request_target`).
2. Point only the credential-bearing jobs (`snapshot-release`, both e2e legs)
   at the `ci-tests` environment, and restrict them so they never run on
   `push` to main.
3. Re-register the required status check as the secret-free aggregator job
   (`passed`), not a credential-bearing job.

Do this out of order and CI deadlocks for everyone.
