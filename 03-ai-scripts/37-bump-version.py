#!/usr/bin/env python3
"""
37-bump-version.py - Autonomous SemVer Version Bumper & Manifest Synchronizer

Bumps version strings across repository manifests, documentation, and changelogs:
  1. Resolves canonical version from version.json, version/info.go, or package.json.
  2. Calculates next SemVer (minor default per Rule 0, patch resets to 0).
  3. Updates version.json, package.json, version/info.go, readme.md, and CHANGELOG.md.
  4. Triggers npm run sync if defined in package.json to regenerate artifacts.
  5. Adheres strictly to repository-aware configuration and file conventions.

Usage:
  python 03-ai-scripts/37-bump-version.py
  python 03-ai-scripts/37-bump-version.py --tier patch
  python 03-ai-scripts/37-bump-version.py --tier minor --scope "Feature release"
  python 03-ai-scripts/37-bump-version.py --tier major --scope "Breaking change"
  python 03-ai-scripts/37-bump-version.py --version 2.323.0
  python 03-ai-scripts/37-bump-version.py --dry-run
"""

import argparse
import datetime
import json
import os
import re
import subprocess
import sys
from pathlib import Path

# Repository root discovery
REPO_ROOT = Path(__file__).resolve().parent.parent

# Canonical version files
VERSION_JSON = REPO_ROOT / "version.json"
PACKAGE_JSON = REPO_ROOT / "package.json"
VERSION_INFO_GO = REPO_ROOT / "version" / "info.go"
README_MD = REPO_ROOT / "readme.md"
CHANGELOG_MD = REPO_ROOT / "CHANGELOG.md"
CHANGELOG_LOWER = REPO_ROOT / "changelog.md"
SPEC19_CHANGELOG = REPO_ROOT / "02-spec" / "19-main-worker-service" / "98-changelog.md"
TEMPLATE_VERSION = REPO_ROOT / "prompt-version.template.json"


def run_cmd(cmd, cwd=None, check=True, capture_output=True):
    """Executes a command with cross-platform safety."""
    target_cwd = cwd or str(REPO_ROOT)
    result = subprocess.run(
        cmd,
        cwd=target_cwd,
        shell=False,
        check=check,
        capture_output=capture_output,
        text=True,
    )
    return result


def get_last_commit_sha():
    """Returns the current commit SHA."""
    try:
        res = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=str(REPO_ROOT),
            capture_output=True,
            text=True,
            check=True,
        )
        return res.stdout.strip()
    except Exception:
        return "unknown"


def get_owner_repo():
    """Discovers <owner>/<repo> from git config."""
    try:
        res = subprocess.run(
            ["git", "config", "--get", "remote.origin.url"],
            cwd=str(REPO_ROOT),
            capture_output=True,
            text=True,
            check=True,
        )
        url = res.stdout.strip()
        if "github.com/" in url:
            return url.split("github.com/")[1].rstrip(".git").strip()
        elif "github.com:" in url:
            return url.split("github.com:")[1].rstrip(".git").strip()
    except Exception:
        pass
    return "alimtvnetwork/movie-cli-v8"


def read_canonical_version():
    """Reads current SemVer from version/info.go, version.json, or package.json."""
    if VERSION_INFO_GO.is_file():
        try:
            with open(VERSION_INFO_GO, "r", encoding="utf-8") as f:
                match = re.search(r'Version\s*=\s*"v?([^"]+)"', f.read())
                if match:
                    return match.group(1).strip()
        except Exception:
            pass

    if VERSION_JSON.is_file():
        try:
            with open(VERSION_JSON, "r", encoding="utf-8") as f:
                data = json.load(f)

            raw_ver = data.get("Version") or data.get("version")
            if raw_ver:
                return str(raw_ver).strip().lstrip("v")
        except Exception:
            pass

    if PACKAGE_JSON.is_file():
        try:
            with open(PACKAGE_JSON, "r", encoding="utf-8") as f:
                data = json.load(f)

            raw_ver = data.get("version")
            if raw_ver and raw_ver != "0.0.0":
                return str(raw_ver).strip().lstrip("v")
        except Exception:
            pass

    return "2.322.1"


def parse_semver(ver_str):
    """Parses X.Y.Z into a tuple of ints (major, minor, patch)."""
    clean_ver = ver_str.lstrip("v")
    match = re.match(r"^(\d+)\.(\d+)\.(\d+)$", clean_ver)
    if not match:
        raise ValueError(f"Invalid SemVer format: '{ver_str}' (expected X.Y.Z)")

    return int(match.group(1)), int(match.group(2)), int(match.group(3))


def calculate_next_version(current_ver, tier):
    """Calculates next SemVer based on tier (Rule 0: default minor, patch resets to 0)."""
    major, minor, patch = parse_semver(current_ver)

    if tier == "patch":
        patch += 1
    elif tier == "minor":
        minor += 1
        patch = 0
    elif tier == "major":
        major += 1
        minor = 0
        patch = 0
    else:
        raise ValueError(f"Unknown bump tier: '{tier}'. Expected patch, minor, or major.")

    return f"{major}.{minor}.{patch}"


def update_version_json(next_version, today_str, dry_run=False):
    """Updates or initializes version.json."""
    owner_repo = get_owner_repo()
    last_sha = get_last_commit_sha()

    data = {
        "Version": next_version,
        "version": next_version,
        "Title": "Movie CLI",
        "RepoSlug": "movie-cli-v8",
        "RepoUrl": f"https://github.com/{owner_repo}",
        "LastCommitSha": last_sha,
        "Description": "Personal movie & TV show library manager — from the terminal.",
        "releaseDate": today_str,
        "changelog": {
            "file_path": "CHANGELOG.md",
            "format": "keep-a-changelog",
        },
        "Authors": [
            {
                "Name": "Md. Alim Ul Karim",
                "Urls": [f"https://github.com/{owner_repo.split('/')[0]}"],
                "Role": "PrimaryAuthor",
                "Background": "Founder and lead developer of Movie CLI.",
            }
        ],
    }

    if VERSION_JSON.is_file():
        try:
            with open(VERSION_JSON, "r", encoding="utf-8") as f:
                existing = json.load(f)
            existing["version"] = next_version
            if "Version" in existing:
                existing["Version"] = next_version
            existing["releaseDate"] = today_str
            existing["LastCommitSha"] = last_sha
            data = existing
        except Exception:
            pass

    if dry_run:
        print(f"[DRY RUN] Would update version.json to {next_version} ({today_str})")
        return

    with open(VERSION_JSON, "w", encoding="utf-8", newline="\n") as f:
        json.dump(data, f, indent=2)
        f.write("\n")

    print(f"[*] Updated version.json -> {next_version}")


def update_version_info_go(next_version, dry_run=False):
    """Updates Version in version/info.go."""
    if not VERSION_INFO_GO.is_file():
        return

    with open(VERSION_INFO_GO, "r", encoding="utf-8") as f:
        content = f.read()

    new_content = re.sub(
        r'Version\s*=\s*"v[^"]+"',
        f'Version   = "v{next_version}"',
        content,
    )

    if new_content == content:
        return

    if dry_run:
        print(f"[DRY RUN] Would update version/info.go to v{next_version}")
        return

    with open(VERSION_INFO_GO, "w", encoding="utf-8", newline="\n") as f:
        f.write(new_content)

    run_cmd(["gofmt", "-w", str(VERSION_INFO_GO)], check=False)
    print(f"[*] Updated and formatted version/info.go -> v{next_version}")


def update_package_json(next_version, dry_run=False):
    """Updates version in package.json."""
    if not PACKAGE_JSON.is_file():
        return

    with open(PACKAGE_JSON, "r", encoding="utf-8") as f:
        data = json.load(f)

    data["version"] = next_version

    if dry_run:
        print(f"[DRY RUN] Would update package.json to {next_version}")
        return

    with open(PACKAGE_JSON, "w", encoding="utf-8", newline="\n") as f:
        json.dump(data, f, indent=2)
        f.write("\n")

    print(f"[*] Updated package.json -> {next_version}")


def update_template_version(next_version, dry_run=False):
    """Updates prompt-version.template.json if present."""
    if not TEMPLATE_VERSION.is_file():
        return

    with open(TEMPLATE_VERSION, "r", encoding="utf-8") as f:
        data = json.load(f)

    data["version"] = next_version

    if dry_run:
        print(f"[DRY RUN] Would update prompt-version.template.json to {next_version}")
        return

    with open(TEMPLATE_VERSION, "w", encoding="utf-8", newline="\n") as f:
        json.dump(data, f, indent=2)
        f.write("\n")

    print(f"[*] Updated prompt-version.template.json -> {next_version}")


def update_readme_pins(current_ver, next_version, dry_run=False):
    """Pins new version in readme.md and ensures install snippets."""
    if not README_MD.is_file():
        return

    with open(README_MD, "r", encoding="utf-8") as f:
        content = f.read()

    owner_repo = get_owner_repo()

    # Update legacy pinned version v2.130.0 -> v{next_version}
    new_content = re.sub(r"v2\.130\.0", f"v{next_version}", content)

    if current_ver:
        new_content = re.sub(rf"\bv?{re.escape(current_ver)}\b", f"v{next_version}", new_content)

    install_snippet = f"""### Install Movie CLI v{next_version}

**Windows (PowerShell):**
```powershell
irm https://github.com/{owner_repo}/releases/download/v{next_version}/install.ps1 | iex
```

**Linux / macOS (Bash):**
```bash
curl -fsSL https://github.com/{owner_repo}/releases/download/v{next_version}/install.sh | bash
```
"""

    if f"Install Movie CLI v{next_version}" not in new_content:
        if "### Pinned to a specific release" in new_content:
            new_content = new_content.replace(
                "### Pinned to a specific release",
                f"{install_snippet}\n### Pinned to a specific release",
            )

    if new_content == content:
        return

    if dry_run:
        print(f"[DRY RUN] Would update version references in readme.md: -> v{next_version}")
        return

    with open(README_MD, "w", encoding="utf-8", newline="\n") as f:
        f.write(new_content)

    print(f"[*] Updated readme.md version pins -> v{next_version}")


def update_changelogs(next_version, scope, today_str, bullets=None, dry_run=False):
    """Prepends release entries to CHANGELOG.md / changelog.md and writes release notes."""
    owner_repo = get_owner_repo()
    cl_target = CHANGELOG_MD if CHANGELOG_MD.is_file() else CHANGELOG_LOWER

    entry_bullets = []
    if bullets:
        entry_bullets.extend(bullets)
    elif scope:
        entry_bullets.append(f"- **{scope}**")
    else:
        entry_bullets.append(f"- **Release v{next_version}**")

    bullets_md = "\n".join(entry_bullets)

    entry_header = f"""## v{next_version}

### Added / Changed
{bullets_md}

### Quick Install Movie CLI v{next_version}

**Windows (PowerShell):**
```powershell
irm https://github.com/{owner_repo}/releases/download/v{next_version}/install.ps1 | iex
```

**Linux / macOS (Bash):**
```bash
curl -fsSL https://github.com/{owner_repo}/releases/download/v{next_version}/install.sh | bash
```

"""

    if cl_target.is_file():
        with open(cl_target, "r", encoding="utf-8") as f:
            cl_content = f.read()

        if f"## v{next_version}" not in cl_content:
            if dry_run:
                print(f"[DRY RUN] Would prepend changelog entry to {cl_target.name} for v{next_version}")
            else:
                if "# Changelog\n" in cl_content:
                    cl_content = cl_content.replace("# Changelog\n", f"# Changelog\n\n{entry_header}", 1)
                else:
                    cl_content = f"# Changelog\n\n{entry_header}{cl_content}"

                with open(cl_target, "w", encoding="utf-8", newline="\n") as f:
                    f.write(cl_content)

                print(f"[*] Prepended changelog entry in {cl_target.name} -> v{next_version}")

    # Generate dedicated release notes artifact for GitHub Release
    release_dir = REPO_ROOT / ".ai-memory" / "release"
    release_dir.mkdir(parents=True, exist_ok=True)
    notes_file = release_dir / f"release-notes-v{next_version}.md"
    notes_content = f"""## Quick Install v{next_version}

### Windows (PowerShell)

```powershell
irm https://github.com/{owner_repo}/releases/download/v{next_version}/install.ps1 | iex
```

### Unix / Linux / macOS (Bash)

```bash
curl -fsSL https://github.com/{owner_repo}/releases/download/v{next_version}/install.sh | bash
```

---

## What's Changed in v{next_version}

### Added / Changed — {scope}

{bullets_md}
"""
    if dry_run:
        print(f"[DRY RUN] Would write release notes to {notes_file.relative_to(REPO_ROOT)}")
    else:
        with open(notes_file, "w", encoding="utf-8", newline="\n") as f:
            f.write(notes_content)

        print(f"[*] Generated release notes -> {notes_file.relative_to(REPO_ROOT)}")

    if SPEC19_CHANGELOG.is_file():
        with open(SPEC19_CHANGELOG, "r", encoding="utf-8") as f:
            s19_content = f.read()

        if f"v{next_version}" not in s19_content:
            s19_entry = f"## v{next_version} — {today_str} ({scope})\n\n**Scope:** Version bump. {scope}.\n\n---\n\n"
            if dry_run:
                print(f"[DRY RUN] Would prepend entry to {SPEC19_CHANGELOG.name}")
            else:
                s19_content = f"{s19_entry}{s19_content}"
                with open(SPEC19_CHANGELOG, "w", encoding="utf-8", newline="\n") as f:
                    f.write(s19_content)

                print(f"[*] Prepended entry in {SPEC19_CHANGELOG.relative_to(REPO_ROOT)} -> v{next_version}")


def run_repo_sync_if_available(dry_run=False):
    """Executes `npm run sync` if defined in package.json to regenerate spec trees and manifests."""
    if not PACKAGE_JSON.is_file():
        return

    try:
        with open(PACKAGE_JSON, "r", encoding="utf-8") as f:
            pkg = json.load(f)

        scripts = pkg.get("scripts", {})
        if "sync" in scripts:
            if dry_run:
                print("[DRY RUN] Would run: npm run sync")
                return

            print("[*] Running npm run sync to regenerate spec trees and manifests...")
            run_cmd(["npm", "run", "sync"], check=False)
            print("[*] Completed npm run sync.")
    except Exception as e:
        print(f"[!] Warning running npm run sync: {e}")


def execute_bump(tier="minor", explicit_version=None, scope=None, bullets=None, dry_run=False):
    """Main bump execution logic."""
    current_ver = read_canonical_version()

    if explicit_version:
        next_ver = explicit_version.lstrip("v")
    else:
        next_ver = calculate_next_version(current_ver, tier)

    today_str = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d")
    bump_scope = scope or f"Routine release v{next_ver}"

    print(f"[*] Previous version: v{current_ver}")
    print(f"[*] Next version:     v{next_ver} (Tier: {tier})")

    update_version_json(next_ver, today_str, dry_run=dry_run)
    update_version_info_go(next_ver, dry_run=dry_run)
    update_package_json(next_ver, dry_run=dry_run)
    update_template_version(next_ver, dry_run=dry_run)
    update_readme_pins(current_ver, next_ver, dry_run=dry_run)
    update_changelogs(next_ver, bump_scope, today_str, bullets=bullets, dry_run=dry_run)
    run_repo_sync_if_available(dry_run=dry_run)

    print(f"[OK] Successfully bumped version to {next_ver}")
    return next_ver


def parse_arguments():
    """Configures CLI argument parser."""
    parser = argparse.ArgumentParser(
        description="37-bump-version: Repository-aware SemVer version bumper & manifest synchronizer."
    )
    parser.add_argument(
        "-t",
        "--tier",
        choices=["patch", "minor", "major"],
        default="minor",
        help="SemVer bump tier (default: minor per Rule 0)",
    )
    parser.add_argument(
        "-v",
        "--version",
        dest="explicit_version",
        default=None,
        help="Explicit SemVer string (overrides --tier)",
    )
    parser.add_argument(
        "-s",
        "--scope",
        default=None,
        help="One-line description/scope of the release",
    )
    parser.add_argument(
        "-b",
        "--bullet",
        dest="bullets",
        action="append",
        default=None,
        help="Individual changelog bullet point (can specify multiple times)",
    )
    parser.add_argument(
        "--bullets-file",
        dest="bullets_file",
        default=None,
        help="Path to file containing markdown bullet points",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simulate the bump without modifying files",
    )

    return parser.parse_args()


def main():
    """CLI entrypoint."""
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    if hasattr(sys.stderr, "reconfigure"):
        sys.stderr.reconfigure(encoding="utf-8")

    args = parse_arguments()

    bullets = args.bullets or []
    if args.bullets_file:
        bf_path = Path(args.bullets_file)
        if bf_path.is_file():
            lines = [l.strip() for l in bf_path.read_text(encoding="utf-8").splitlines() if l.strip()]
            bullets.extend(lines)

    resolved_bullets = bullets if bullets else None

    execute_bump(
        tier=args.tier,
        explicit_version=args.explicit_version,
        scope=args.scope,
        bullets=resolved_bullets,
        dry_run=args.dry_run,
    )


if __name__ == "__main__":
    main()
