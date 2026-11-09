---
name: deps-update
description: "Apply a YAML update report (sections `images`, `dependencies` and `charts`, the automation-hub/Renovate dependency-report format) to the local devidence-dev repos. Phase 1 updates Docker images; phase 2, after the user confirms the images were pushed, updates dependencies (Go via go get, Python via uv, npm/bun, workflows, charts) using docker run. Use when the user pastes a YAML with `scanned_at` and any of `images:`, `dependencies:` or `charts:`, or asks to update dependencies/images from the report."
---

# deps-update

Apply the update report to `${DEPS_UPDATE_REPOS_ROOT:-$HOME/repositories}/<repo>`. Local changes only: **never** `git commit` or `git push`. You may run `git add` and show the diff.

## Input

Pasted YAML or a file path:

```yaml
scanned_at: "..."
images:        # phase 1
  - repo: "devidence-dev/<repo>"
    fork: false
    files:
      - path: "<path inside the repo>"
        updates:
          - { name, current, new, type: major|minor|patch }
dependencies:  # phase 2, same shape
charts:        # phase 2, same shape
```

## Preparation

1. Map `devidence-dev/<repo>` to `${DEPS_UPDATE_REPOS_ROOT:-$HOME/repositories}/<repo>`. If it does not exist, report it and continue. Do not clone.
2. `fork: true`: skip the repo and report it.
3. Per repo: `git status --short`. If a file to edit already has uncommitted changes, warn and ask before editing it.
4. Resolve duplicates of the same `name` in one file (e.g. node major + minor) with the **Version selection** rules before editing.

## Version selection

- **patch / minor**: apply directly.
- **node**: pick the newest LTS among the report's options. Check `https://nodejs.org/dist/index.json` (field `lts` not `false`). If the new major is already LTS, use it; otherwise keep the minor/patch on the current line.
- **other majors**: do not apply without asking. Present them together, each with a recommendation (breaking-change risk), and apply only the ones the user approves. If a minor/patch alternative exists for the same `name`, apply that meanwhile.
  - A Go major (`v1` to `v2`) changes the module path (`.../v2`) and imports: flag it.
- **Consistency**: the same tool must end at the same version across the whole repo (e.g. node in `Dockerfile`, `.nvmrc` and `package.json` `engines`; `oven/bun` in `Dockerfile` and `.devcontainer`), even if one report section lists it and another does not.

## Phase 1: images

Edit only the version text, keeping the existing format:

- **Dockerfile** `FROM img:<ver><suffix>`: change only the version, keep the suffix (`-alpine`, `-slim-trixie`) and `AS <stage>`. If there is an `@sha256:...`, resolve the new digest with `docker buildx imagetools inspect img:<tag>` and replace it.
- **devcontainer.json**: `image` same as Dockerfile. `features` (`ghcr.io/devcontainers/features/x:<ver>`): change the tag.
- **k8s/helm YAML** (`values.yaml`, manifests): change `tag:` or `image: img:<ver>`, keeping the `v` prefix if present.

When phase 1 is done:

1. Show `git diff --stat` and the diff per repo.
2. Optional check if the user asks: `docker build` of the modified Dockerfile.
3. **Stop.** Tell the user to commit/push the images and wait for their confirmation before phase 2.

## Phase 2: dependencies

All tools run in a container, never locally. Base for all:

```bash
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$PWD":/app -w /app <image> <cmd>
```

Use as `<image>` the version left in the repo's Dockerfile after phase 1 (golang, node, bun, python). If the repo has no Dockerfile with that image, use the version from the `go` directive / `.nvmrc` / `requires-python`.

### Go (`go.mod`)

```bash
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
  -v "$PWD":/app -w /app golang:<ver>-alpine \
  sh -c 'go get <mod>@<new> [<mod>@<new> ...] && go mod tidy && go build ./...'
```

- One single call with all of the repo's modules.
- Exact version from the report (`go get mod@ver`), not bare `go get -u`: `-u` also bumps transitive deps outside the report.
- Do not touch the `go` directive or `toolchain` unless the report asks for it.

### Python (`pyproject.toml`)

- Dependency with a range (e.g. `uv_build>=0.12.0,<0.13.0`): move the range to the new version keeping its shape (`>=0.13.7,<0.14.0`).
- Then regenerate the lock:

```bash
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e UV_CACHE_DIR=/tmp/uv \
  -v "$PWD":/app -w /app ghcr.io/astral-sh/uv:<ver>-debian-slim \
  uv lock --upgrade-package <name>
```

- `<ver>`: the uv/uv_build version from the report, or the repo's.
- `[build-system]` entries are not in `uv.lock`; for those, editing the range is enough.

### Node / Bun (`package.json`)

- Edit the version in `package.json` keeping the prefix (`^`, `~`, exact). `engines.node` and `.nvmrc` follow the node rule (keep the `v` in `.nvmrc` if present).
- Regenerate the lock based on which file exists:
  - `bun.lock`: `oven/bun:<ver>` with `bun install --ignore-scripts`
  - `package-lock.json`: `node:<ver>-alpine` with `npm install --ignore-scripts`

### Text edits only

- **Reusable workflows / actions** (`uses: owner/repo/...@<ref>`): respect the ref style. If it is floating (`@v1`) and the update is a minor/patch of the same major, nothing changes: report it as "already covered". If it is a major, the majors rule applies and the ref becomes `@v2` (floating if it was floating).
- **ArgoCD/Helm** (`targetRevision`, chart `version`): change the value keeping the quotes. Entries of the `charts:` section use this rule: in the file at `path`, change `targetRevision` only in the source block whose `chart:` equals `name`, because one file can hold several sources, keeping the quotes; the majors rule applies.

## Wrap-up

1. `git add` the touched files per repo (no commit).
2. Final summary per repo:
   - Applied (`name current -> new`).
   - Skipped with reason: major not approved, already covered by floating tag, fork, missing repo, build failure.
   - Docker commands that failed, with the decisive error line.
3. Remind the user that commit/push is theirs to do.
