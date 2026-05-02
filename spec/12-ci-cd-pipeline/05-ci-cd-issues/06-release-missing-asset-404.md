# 06 — Release Missing Asset (HTTP 404 on install)

## Symptom

User runs the documented one-liner from a GitHub Release page:

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.97.0/install.ps1 | iex
```

and sees:

```
  movie installer (v2.97.0)
  github.com/alimtvnetwork/movie-cli-v8

  Downloading movie-v2.97.0-windows-amd64.zip (v2.97.0)...
  Download failed: Not Found
```

The `install.ps1` itself downloads fine (it's hosted on the release), but the binary archive 404s.

## Root cause

The release was published with a **partial asset set**. Inspection of the `v2.97.0` release showed:

| Expected asset                            | Present? |
|-------------------------------------------|----------|
| `movie-v2.97.0-windows-amd64.zip`         | **NO**   |
| `movie-v2.97.0-windows-arm64.zip`         | **NO**   |
| `movie-v2.97.0-linux-amd64.tar.gz`        | **NO**   |
| `movie-v2.97.0-linux-arm64.tar.gz`        | yes      |
| `movie-v2.97.0-darwin-amd64.tar.gz`       | yes      |
| `movie-v2.97.0-darwin-arm64.tar.gz`       | yes      |
| `checksums.txt`                           | yes      |
| `install.ps1` / `install.sh`              | yes      |

The Windows builds and one Linux build were never uploaded — most likely the `Embed icon for Windows builds` step (`go-winres`) failed silently, or an individual `go build` step crashed for a single target while the workflow's continue-on-error behavior let the rest of the pipeline finish.

The `softprops/action-gh-release@v2` step ran with `files: dist/*` and uploaded **whatever was in `dist/`** — it has no concept of "this release is incomplete."

## Trigger

Any code path in `.github/workflows/release.yml` that lets the workflow reach the `Create GitHub Release` step with fewer than 6 archives in `dist/`.

## Fix pattern

### 1. CI — fail fast on incomplete asset set

Add a verification step **after** `Compress and checksum` and **before** any install-script generation or upload. The step enumerates the 6 expected filenames and exits non-zero if any are missing:

```yaml
- name: Verify all 6 archives are present
  working-directory: dist
  run: |
    VERSION="${{ steps.version.outputs.version }}"
    REQUIRED=(
      "movie-${VERSION}-windows-amd64.zip"
      "movie-${VERSION}-windows-arm64.zip"
      "movie-${VERSION}-linux-amd64.tar.gz"
      "movie-${VERSION}-linux-arm64.tar.gz"
      "movie-${VERSION}-darwin-amd64.tar.gz"
      "movie-${VERSION}-darwin-arm64.tar.gz"
    )
    MISSING=()
    for f in "${REQUIRED[@]}"; do
      [[ -f "$f" ]] || MISSING+=("$f")
    done
    if [[ ${#MISSING[@]} -gt 0 ]]; then
      echo "::error::Release is missing required archives:"
      for f in "${MISSING[@]}"; do echo "::error::  $f"; done
      exit 1
    fi
```

This fires BEFORE the GitHub Release is created, so a broken build never goes public.

### 2. install.ps1 — actionable error on 404

The generated `install.ps1` previously printed `Download failed: <raw exception>` which gave the user no path forward. Update the catch block to detect HTTP 404 specifically and print:

- The exact URL that 404'd (so the user can confirm in a browser)
- An explanation that this is a publisher-side issue, not their machine
- Two workarounds: (a) try a different release tag, (b) build from source via `git clone … && ./run.ps1`

See lines 171-198 of `.github/workflows/release.yml` for the implementation.

### 3. Repo-root `install.ps1` — keep repo URL in sync

The legacy `install.ps1` at the repo root (used for `git clone + build` flow) had a stale `RepoUrl` pointing to `movie-cli-v8.git`, which no longer exists. The build-from-source workaround printed by the 404 handler dies immediately. Fixed to `movie-cli-v8.git`.

**Prevention rule**: any time the repo is renamed/forked to a new `-v<N>` suffix, search for the old name with:

```bash
grep -rn 'movie-cli-v[0-9]' --include='*.ps1' --include='*.sh' --include='*.md' --include='*.go'
```

and update every hit.

## Acceptance criteria

- GIVEN a release where one `go build` target fails
- WHEN the `Verify all 6 archives are present` step runs
- THEN the workflow fails with an explicit list of missing files BEFORE the GitHub Release is created
- AND no partial release is ever published

- GIVEN a user runs `irm .../install.ps1 | iex` against an existing partial release
- WHEN the binary download returns HTTP 404
- THEN they see the exact missing URL, an explanation, and two recovery commands

## History

- **v2.97.0** — first observed; user reported `Download failed: Not Found`. Investigation showed Windows + linux-amd64 archives were never uploaded.
- **v2.128.4** — verification step added to `release.yml`; install.ps1 generator gained 404-aware error handler; repo-root `install.ps1` `RepoUrl` corrected from `-v3` → `-v5`. This issue file authored.

## Related

- `spec/13-self-update-app-update/05-release-distribution.md` — asset matrix and acceptance criteria
- `spec/03-general/05-install-latest-sibling-repo.md` — bootstrap.ps1 fallback algorithm (probes sibling repos when current is broken)
