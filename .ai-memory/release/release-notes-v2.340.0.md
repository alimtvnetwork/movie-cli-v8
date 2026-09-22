## Quick Install v2.340.0

### Windows (PowerShell)

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.340.0/install.ps1 | iex
```

### Unix / Linux / macOS (Bash)

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.340.0/install.sh | bash
```

---

## What's Changed in v2.340.0

### Added / Changed — Add movie cd navigation, folder aliasing, ls root folders, and web ui shortcuts

Implement movie cd with GitMap-style jump targets, clean path stdout, and shell mcd setup
Implement movie alias (list, set, remove, show, suggest) for folder shortcuts
Enhance movie ls with --folders card showing root folders, counts, quick jump, and web ui commands
Enhance movie ui to launch scoped to specific folder, alias, or index from anywhere
