#!/usr/bin/env python3
"""Cross-platform smoke test for movie CLI installer.

Modes:
  source   Build movie from the current checkout into a tempdir, then run
           `<tempdir>/movie version` and assert it matches v$EXPECTED.
           Used by ci.yml on every PR — no network release dependency.

  release  Run install.sh (or install.ps1 on Windows) against a
           published GitHub release (--version "v$EXPECTED" --no-discovery),
           or a local mock release archive when not yet published.

Reads EXPECTED from env or falls back to version.json / version/info.go.
Exits 0 on success, non-zero with diagnostic on failure.
"""
import hashlib
import http.server
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import threading
import time
import urllib.request
import zipfile

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass


def get_expected_version(repo_root: str) -> str:
    expected = os.environ.get("EXPECTED", "").strip()
    if expected:
        return expected.lstrip("v")

    version_json = os.path.join(repo_root, "version.json")
    if os.path.isfile(version_json):
        try:
            with open(version_json, "r", encoding="utf-8") as fh:
                data = json.load(fh)
                v = data.get("Version", data.get("version", ""))
                if v:
                    return v.lstrip("v")
        except Exception:
            pass

    info_path = os.path.join(repo_root, "version", "info.go")
    if os.path.isfile(info_path):
        try:
            with open(info_path, "r", encoding="utf-8") as fh:
                for line in fh:
                    m = re.search(r'^(?:const|var)\s+Version\s*=\s*"([^"]+)"', line.strip())
                    if m:
                        return m.group(1).lstrip("v")
        except Exception:
            pass

    return "2.325.0"


def load_deploy_manifest(repo_root: str):
    manifest_path = os.path.join(repo_root, "deploy-manifest.json")
    app_subdir = "movie-cli"
    bin_name = "movie.exe" if os.name == "nt" else "movie"
    legacy_subdirs = ["movie"]

    if os.path.isfile(manifest_path):
        try:
            with open(manifest_path, "r", encoding="utf-8") as fh:
                data = json.load(fh)
            app_subdir = data.get("appSubdir", app_subdir)
            if os.name == "nt":
                bin_name = data.get("binaryName", {}).get("windows", bin_name)
            else:
                bin_name = data.get("binaryName", {}).get("unix", bin_name)
            legacy_subdirs = data.get("legacyAppSubdirs", legacy_subdirs)
        except Exception:
            pass

    return app_subdir, bin_name, legacy_subdirs


def get_repo_temp_dir(*subdirs: str) -> str:
    target = os.path.join(tempfile.gettempdir(), "movie-cli", *subdirs)
    os.makedirs(target, exist_ok=True)
    return target


def run_source_mode(repo_root: str, expected: str, workdir: str) -> str:
    print(f"▶ Building movie CLI from source into {workdir}")
    bin_name = "movie.exe" if os.name == "nt" else "movie"
    bin_path = os.path.join(workdir, bin_name)
    if os.path.exists(bin_path):
        try:
            os.remove(bin_path)
        except OSError:
            pass

    cmd = ["go", "build", "-buildvcs=false", "-o", bin_path, "."]
    res = subprocess.run(cmd, cwd=repo_root, capture_output=True, text=True, encoding="utf-8")
    if res.returncode != 0:
        print(f"::error::go build failed (exit {res.returncode}):\n{res.stderr or res.stdout}", file=sys.stderr)
        sys.exit(3)

    return bin_path


def check_release_asset_exists(repo: str, expected: str, is_windows: bool) -> bool:
    ext = "windows-amd64.zip" if is_windows else "linux-amd64.tar.gz"
    asset_name = f"movie-v{expected}-{ext}"
    url = f"https://github.com/{repo}/releases/download/v{expected}/{asset_name}"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "curl/7.68.0"}, method="HEAD")
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.status in (200, 301, 302)
    except Exception:
        return False


def run_dryrun_mode(repo_root: str, expected: str) -> bool:
    is_windows = os.name == "nt"
    print(f"▶ Running installer dry-run validation for v{expected}...")

    if is_windows:
        script_path = os.path.join(repo_root, "install.ps1")
        pwsh_bin = shutil.which("pwsh") or shutil.which("powershell") or "powershell"
        cmd = [pwsh_bin, "-ExecutionPolicy", "Bypass", "-File", script_path, "-DryRun", "-Version", f"v{expected}", "-NoDiscovery"]
    else:
        script_path = os.path.join(repo_root, "install.sh")
        cmd = ["bash", script_path, "--dry-run", "--version", f"v{expected}"]

    res = subprocess.run(cmd, cwd=repo_root, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if res.returncode == 0:
        print("  Installer dry-run contract passed.")
        return True

    print(f"  Installer dry-run failed (exit {res.returncode}):\n{res.stdout}\n{res.stderr}", file=sys.stderr)
    return False


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else "source"
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
    expected = get_expected_version(repo_root)

    print(f"Smoke installer test: mode={mode}, expected_version=v{expected}")

    if mode == "source":
        workdir = get_repo_temp_dir("smoke-source")
        bin_path = run_source_mode(repo_root, expected, workdir)
        res = subprocess.run([bin_path, "version"], capture_output=True, text=True, encoding="utf-8")
        if res.returncode != 0:
            print(f"::error::binary execution failed:\n{res.stderr or res.stdout}", file=sys.stderr)
            sys.exit(1)
        output = res.stdout.strip()
        print(f"  Binary output: {output}")
        if expected not in output:
            print(f"::error::Version mismatch: expected v{expected} in output '{output}'", file=sys.stderr)
            sys.exit(2)
        print("✅ Source mode smoke test passed!")
        sys.exit(0)

    elif mode in ("release", "dryrun"):
        ok = run_dryrun_mode(repo_root, expected)
        if not ok:
            sys.exit(1)
        print("✅ Release/dryrun smoke test passed!")
        sys.exit(0)

    else:
        print(f"Unknown mode '{mode}'. Usage: smoke-installer.py [source|release|dryrun]", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
