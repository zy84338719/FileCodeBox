## Contributing and CI guidelines

This document explains how our GitHub Actions are gated and how to trigger heavier CI jobs (cross-platform builds, Docker tests, security scans) intentionally.

### CI design principles
- Lightweight checks (unit tests, linters) run on `push` / `pull_request` for `main` and `develop` branches.
- Heavy tasks (cross-platform builds, Docker image builds & push, release packaging) only run on tag pushes (e.g. `v1.10.2`) or when explicitly requested.

This reduces wasted CI minutes and avoids building/publishing artifacts for every code merge.

### How to trigger heavy CI jobs
There are multiple intentional ways to trigger full/heavy workflows:

- Push a semantic version tag (recommended for releases):
  ```bash
  git tag v1.10.2
  git push origin v1.10.2
  ```
  Tag pushes trigger the `build.yml` / `release.yml` flows which do cross-platform builds and create the Release.

- Use commit message flags for ad-hoc full CI when pushing branches:
  - `[docker-test]` — triggers Docker Test job
  - `[security-scan]` — triggers Security Scan job
  Example:
  ```bash
  git commit -m "feat: add X [docker-test]"
  git push origin feature/xxx
  ```

- Use PR labels to request full CI for a pull request (team members with label permissions can add labels):
  - `run-full-ci` — triggers Docker Test
  - `run-security` — triggers Security Scan

- Manual dispatch from GitHub Actions UI (`workflow_dispatch`) — allowed for `build.yml` and can be used to run ad-hoc full builds.

### Release flow (recommended)
1. Finish work on a branch, open PR and get reviews.
2. Merge to `main` when ready (this runs lightweight CI only).
3. Create a version tag on the merge commit and push the tag:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
4. The tag push will run the release pipeline (build artifacts, create release, push Docker images).

### Generated files and embeds
- Some files under `uploads/` or other `generated/` directories may be produced by code generation tools. Do not commit generated artifacts unless they are intentionally tracked.
- If a file uses `//go:embed` to include generated assets, ensure the assets exist in the repository or exclude the embed file from normal builds (recommended: keep generated assets out of VCS and generate them in CI when needed).

Recommended `.gitignore` additions for generated assets (add if appropriate):
```
# generated embed or build artifacts
uploads/
dist/
build/
```

### Debugging CI triggers
- To see why a job ran, open the Actions page, find the workflow run and view the `Event` and `Jobs` details. The `Event` will indicate whether it was a `push`, `pull_request`, `workflow_dispatch`, or `tag` event.
- If you expected a heavy job but it didn't run, verify:
  - You pushed a tag (tags trigger build/release flows).
  - The commit message includes the required token (e.g., `[docker-test]`).
  - A PR contains the appropriate label (`run-full-ci` or `run-security`).

### Contact / ownership
- CI workflow files are located under `.github/workflows/`. If you want to change gating logic, please open a PR and tag the maintainers.

---
Thank you for contributing — keeping heavy CI runs intentional saves time and cost for the whole team.
