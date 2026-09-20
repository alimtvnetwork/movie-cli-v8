#!/usr/bin/env python3
"""check-acronym-naming.py — Enforce Go acronym MixedCaps rule on identifiers only.

Spec: spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md

Forbidden acronyms when followed by another uppercase letter (i.e. inside a
MixedCaps identifier): IMDb, TMDb, API, HTTP, URL, JSON, SQL, HTML, XML.

Comments (// ... and /* ... */) and string/rune literals are stripped before
matching, so prose in doc comments like "OMDb HTTPS URL" is allowed.

Allowlist: short trailing-initialism locals (imdbID, tmdbID, *URL).

Exit codes:
  0  no violations
  1  one or more violations (printed with GitHub `::error::` annotations)
"""
from __future__ import annotations
import os
import re
import sys

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")

ALLOWLIST = re.compile(
    r'\b(imdbID|tmdbID|imgURL|reqURL|posterURL|baseURL|apiURL|fullURL|'
    r'targetURL|rawURL|nextURL|prevURL|sourceURL|destURL|webhookURL|'
    r'callbackURL|redirectURL|releaseURL|downloadURL|assetURL|repoURL|'
    r'cloneURL|htmlURL|avatarURL|profileURL)\b'
)
PATTERN = re.compile(r'\b(IMDb|TMDb|API|HTTP|URL|JSON|SQL|HTML|XML)[A-Z]')

SKIP_DIRS = {'.git', '.release', 'node_modules', '.gitmap'}


def strip_go(src: str) -> str:
    """Replace comments and string/rune literal contents with spaces.

    Newlines are preserved so line numbers in the stripped output match the
    original source 1:1.
    """
    out: list[str] = []
    i, n = 0, len(src)
    state = 'code'
    while i < n:
        c = src[i]
        c2 = src[i:i + 2]
        if state == 'code':
            if c2 == '//':
                out.append('  '); state = 'line_comment'; i += 2; continue
            if c2 == '/*':
                out.append('  '); state = 'block_comment'; i += 2; continue
            if c == '"':
                out.append(' '); state = 'str'; i += 1; continue
            if c == '`':
                out.append(' '); state = 'rstr'; i += 1; continue
            if c == "'":
                out.append(' '); state = 'rune'; i += 1; continue
            out.append(c); i += 1; continue
        if state == 'line_comment':
            if c == '\n':
                out.append('\n'); state = 'code'; i += 1; continue
            out.append('\t' if c == '\t' else ' '); i += 1; continue
        if state == 'block_comment':
            if c2 == '*/':
                out.append('  '); state = 'code'; i += 2; continue
            out.append('\n' if c == '\n' else ('\t' if c == '\t' else ' '))
            i += 1; continue
        if state == 'str':
            if c == '\\' and i + 1 < n:
                out.append('  '); i += 2; continue
            if c == '"':
                out.append(' '); state = 'code'; i += 1; continue
            out.append('\n' if c == '\n' else ' '); i += 1; continue
        if state == 'rstr':
            if c == '`':
                out.append(' '); state = 'code'; i += 1; continue
            out.append('\n' if c == '\n' else ' '); i += 1; continue
        if state == 'rune':
            if c == '\\' and i + 1 < n:
                out.append('  '); i += 2; continue
            if c == "'":
                out.append(' '); state = 'code'; i += 1; continue
            out.append(' '); i += 1; continue
    return ''.join(out)


def scan_file(path: str) -> list[tuple[str, int, str, str]]:
    try:
        src = open(path, encoding='utf-8', errors='replace').read()
    except OSError:
        return []
    stripped = strip_go(src)
    orig_lines = src.splitlines()
    violations: list[tuple[str, int, str, str]] = []
    for lineno, line in enumerate(stripped.splitlines(), 1):
        for m in PATTERN.finditer(line):
            j = m.start()
            while j > 0 and (line[j - 1].isalnum() or line[j - 1] == '_'):
                j -= 1
            k = m.end()
            while k < len(line) and (line[k].isalnum() or line[k] == '_'):
                k += 1
            ident = line[j:k]
            if ALLOWLIST.search(ident):
                continue
            orig = orig_lines[lineno - 1].strip() if lineno - 1 < len(orig_lines) else ''
            violations.append((path, lineno, ident, orig))
    return violations


def main(root: str = '.') -> int:
    violations: list[tuple[str, int, str, str]] = []
    for dirpath, dirs, files in os.walk(root):
        dirs[:] = [d for d in dirs if d not in SKIP_DIRS]
        for f in files:
            if f.endswith('.go'):
                violations.extend(scan_file(os.path.join(dirpath, f)))
    if not violations:
        print("✅ No acronym MixedCaps violations in identifiers.")
        return 0
    print("Acronym MixedCaps violations (identifiers only — comments/strings ignored):")
    print("See spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md")
    for path, lineno, ident, orig in violations:
        print(f"{path}:{lineno}: {ident}    // {orig}")
        print(f"::error file={path},line={lineno}::Acronym MixedCaps in identifier '{ident}'")
    print()
    print("Fix: rename IMDb→Imdb, TMDb→Tmdb, API→Api, HTTP→Http, URL→Url, "
          "JSON→Json, SQL→Sql, HTML→Html, XML→Xml in identifiers.")
    return 1


if __name__ == '__main__':
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else '.'))
