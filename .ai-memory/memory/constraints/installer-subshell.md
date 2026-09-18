---
name: Installer scripts cannot mutate parent shell env
description: Scripts run via curl|bash or irm|iex execute in a subshell and cannot change the parent shell's PATH or env — always print a copy-paste hint instead of silently writing to rc files
type: constraint
---

# Installer Subshell Constraint

Scripts invoked via `curl ... | bash`, `irm ... | iex`, `wget -O- ... | sh`,
or `source <(curl ...)` (except the last) run in a **subshell or piped
process**. They CANNOT mutate the parent interactive shell's environment:

- `export PATH=...` only affects the subshell, gone when the pipe ends
- `[Environment]::SetEnvironmentVariable(..., "User")` only affects *future*
  PowerShell sessions, not the current one
- Writes to `~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish` only apply
  to **new** shells the user opens later

## Required behavior for installers

1. Still write to the rc file / user PATH for future sessions.
2. Be explicit in messaging: e.g. `Added to ~/.zshrc (new shells will pick it up)`.
3. **Always print a copy-pasteable one-liner** the user can run NOW to refresh
   the current shell:
   - bash/zsh: `export PATH="$dir:$PATH"`
   - fish: `fish_add_path $dir`
   - PowerShell: `$env:PATH += ";$dir"`
4. Never claim "movie is now ready to use" without that hint — users will hit
   `command not found` immediately and think the install failed.

## Why
- Unix process isolation: child processes cannot write to a parent's env.
- Windows user-scope env vars are read at session start, not refreshed mid-session.
- Silently writing to rc files looks like the installer did nothing.
