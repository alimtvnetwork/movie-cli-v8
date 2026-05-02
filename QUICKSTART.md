# 🚀 Quickstart

Get the **movie** CLI running in under a minute. For full docs, flags, and
troubleshooting, see [README.md](README.md) and
[spec/03-general/01-install-guide.md](spec/03-general/01-install-guide.md).

---

## 🚀 One-liner install

<!-- INSTALL:BEGIN -->

<!-- Generated from README.md by scripts/sync-install-from-readme.sh — do not edit by hand -->

**🚀 Install in 10 seconds — anyone, any OS:**

<table>
<tr>
<td><b>🪟 Windows · PowerShell</b></td>
<td>

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.ps1 | iex
```

</td>
</tr>
<tr>
<td><b>🐧 macOS · Linux · Bash</b></td>
<td>

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.sh | bash
```

</td>
</tr>
</table>

<sub>Auto-detects your OS &amp; architecture · Installs the latest pre-built binary · Falls back to a source build if no release is published · See <a href="#installation">Installation</a> for flags, pinned versions, and verification.</sub>

<!-- INSTALL:END -->

---

## 🐧 Linux

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.sh | bash
movie version
```

If `movie` is not found:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

---

## 🍎 macOS

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.sh | bash
movie version
```

If `movie` is not found:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

---

## 🪟 Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.ps1 | iex
movie version
```

If `movie` is not found in the current shell:

```powershell
$env:PATH += ";$HOME\bin"
```

---

## ✅ Verify the install

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/verify.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/verify.ps1 | iex
```

Expected: `All required checks passed — movie CLI is ready.`

---

## 🎬 First commands

```bash
movie scan ~/Movies        # scan a folder, match against TMDb
movie ls                   # list your library
movie info "The Matrix"    # show details for one title
movie --help               # full command list
```
