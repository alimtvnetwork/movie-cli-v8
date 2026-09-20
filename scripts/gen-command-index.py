#!/usr/bin/env python3
"""
Single source of truth for the README's "Flat alphabetical command index".

Why this exists
---------------
The index appears twice in README.md:

  1. An HTML <table> for nicely-rendered GitHub viewing.
  2. A fenced ```text``` block for terminals / monospace editors.

Hand-editing both blocks is how drift starts. This script owns the data and
regenerates both regions in place, between explicit marker comments:

  <!-- COMMAND-INDEX:HTML:BEGIN -->   ... <!-- COMMAND-INDEX:HTML:END -->
  <!-- COMMAND-INDEX:TEXT:BEGIN -->   ... <!-- COMMAND-INDEX:TEXT:END -->

It also keeps the six command-section anchors (Scanning & Library, File
Management, History & Undo, Discovery & Organization, Maintenance &
Debugging, Configuration & System) in sync everywhere they appear in the
README. Each section label has exactly ONE canonical GitHub slug, computed
from the label itself; any `(#stale)` link to one of these sections is
rewritten in place when you run the script, and `--check` fails CI if any
such rewrite is still pending. Anchors unrelated to the six sections are
left untouched.

Usage
-----
  python3 scripts/gen-command-index.py            # rewrite README.md in place
  python3 scripts/gen-command-index.py --check    # exit 1 if README is stale

The --check mode is what CI runs to enforce no-drift.
"""
from __future__ import annotations

import argparse
import difflib
import html
import pathlib
import re
import sys

REPO_ROOT = pathlib.Path(__file__).resolve().parent.parent
README = (REPO_ROOT / "readme.md") if (REPO_ROOT / "readme.md").exists() else (REPO_ROOT / "README.md")

# ─── Source of truth ────────────────────────────────────────────────────────
# Each entry: (command, section_label, example_keyword)
#
# - command         : exact CLI invocation, with <placeholders>
# - section_label   : human-readable section name shown in the index. Must be
#                     one of SECTION_LABELS below — its anchor is derived from
#                     the label via _section_slug() so renaming a section in
#                     ONE place updates every link that references it.
# - example_keyword : minimal Ctrl/⌘+F-friendly snippet (trailing space if it's
#                     a prefix, no space if it's already complete)
#
# Adding/renaming a command? Edit ONLY this list, then run the script.
SECTION_LABELS: tuple[str, ...] = (
    "Scanning & Library",
    "File Management",
    "History & Undo",
    "Discovery & Organization",
    "Maintenance & Debugging",
    "Configuration & System",
)

COMMANDS: list[tuple[str, str, str]] = [
    ("movie cd <id>",                              "File Management",         "movie cd "),
    ("movie changelog",                            "Configuration & System",  "movie changelog"),
    ("movie cleanup",                              "Maintenance & Debugging", "movie cleanup"),
    ("movie config",                               "Configuration & System",  "movie config"),
    ("movie config get <key>",                     "Configuration & System",  "movie config get "),
    ("movie config set <key> <value>",             "Configuration & System",  "movie config set "),
    ("movie config set source_folder <path>",      "Configuration & System",  "movie config set source_folder "),
    ("movie config set tmdb_api_key <key>",        "Configuration & System",  "movie config set tmdb_api_key "),
    ("movie db",                                   "Maintenance & Debugging", "movie db"),
    ("movie discover",                             "Discovery & Organization","movie discover"),
    ("movie duplicates",                           "File Management",         "movie duplicates"),
    ("movie export",                               "Maintenance & Debugging", "movie export"),
    ("movie export --format csv --out <file>",     "Maintenance & Debugging", "movie export --format csv "),
    ("movie export --format json --out <file>",    "Maintenance & Debugging", "movie export --format json "),
    ("movie hello",                                "Configuration & System",  "movie hello"),
    ("movie info <id>",                            "Scanning & Library",      "movie info "),
    ("movie info <id> --json",                     "Scanning & Library",      "movie info --json "),
    ("movie logs",                                 "Maintenance & Debugging", "movie logs"),
    ("movie ls",                                   "Scanning & Library",      "movie ls"),
    ("movie ls --genre <name>",                    "Scanning & Library",      "movie ls --genre "),
    ("movie ls --limit <n>",                       "Scanning & Library",      "movie ls --limit "),
    ("movie ls --year <yyyy> --sort <field>",      "Scanning & Library",      "movie ls --year "),
    ("movie move",                                 "File Management",         "movie move"),
    ("movie move --all",                           "File Management",         "movie move --all"),
    ("movie move <id> --to <path>",                "File Management",         "movie move --to "),
    ("movie play <id>",                            "File Management",         "movie play "),
    ("movie play <id> --player <bin>",             "File Management",         "movie play --player "),
    ("movie popout",                               "File Management",         "movie popout"),
    ("movie redo",                                 "History & Undo",          "movie redo"),
    ("movie rename",                               "File Management",         "movie rename"),
    ("movie rename <id>",                          "File Management",         "movie rename "),
    ("movie rename --all --pattern <fmt>",         "File Management",         "movie rename --all --pattern "),
    ("movie rescan",                               "Scanning & Library",      "movie rescan"),
    ("movie rest",                                 "Maintenance & Debugging", "movie rest"),
    ("movie rest --open",                          "Maintenance & Debugging", "movie rest --open"),
    ("movie rest --port <n>",                      "Maintenance & Debugging", "movie rest --port "),
    ("movie scan",                                 "Scanning & Library",      "movie scan"),
    ("movie scan <path>",                          "Scanning & Library",      "movie scan "),
    ("movie scan <path> --dry-run",                "Scanning & Library",      "movie scan --dry-run "),
    ("movie scan <path> --refresh",                "Scanning & Library",      "movie scan --refresh "),
    ("movie search <query>",                       "Scanning & Library",      "movie search "),
    ("movie search <query> --year <yyyy>",         "Scanning & Library",      "movie search --year "),
    ("movie self-uninstall",                       "Configuration & System",  "movie self-uninstall"),
    ("movie stats",                                "Discovery & Organization","movie stats"),
    ("movie stats --by <dimension>",               "Discovery & Organization","movie stats --by "),
    ("movie suggest",                              "Discovery & Organization","movie suggest"),
    ("movie suggest --genre <name> --limit <n>",   "Discovery & Organization","movie suggest --genre "),
    ("movie tag add <id> <tag>",                   "Discovery & Organization","movie tag add "),
    ("movie tag list <id>",                        "Discovery & Organization","movie tag list "),
    ("movie tag list --all",                       "Discovery & Organization","movie tag list --all"),
    ("movie tag remove <id> <tag>",                "Discovery & Organization","movie tag remove "),
    ("movie tag remove <id> --all",                "Discovery & Organization","movie tag remove --all"),
    ("movie undo",                                 "History & Undo",          "movie undo"),
    ("movie undo --id <history-id>",               "History & Undo",          "movie undo --id "),
    ("movie undo --list",                          "History & Undo",          "movie undo --list"),
    ("movie update",                               "Configuration & System",  "movie update"),
    ("movie version",                              "Configuration & System",  "movie version"),
    ("movie watch add <id>",                       "Discovery & Organization","movie watch add "),
    ("movie watch add <id> --priority <level>",    "Discovery & Organization","movie watch add --priority "),
    ("movie watch list",                           "Discovery & Organization","movie watch list"),
    ("movie watch list --sort <field>",            "Discovery & Organization","movie watch list --sort "),
]

HTML_BEGIN = "<!-- COMMAND-INDEX:HTML:BEGIN -->"
HTML_END   = "<!-- COMMAND-INDEX:HTML:END -->"
TEXT_BEGIN = "<!-- COMMAND-INDEX:TEXT:BEGIN -->"
TEXT_END   = "<!-- COMMAND-INDEX:TEXT:END -->"

# Per-section quick-start blocks live under each `#### 📂 [Section]` heading
# in the Command Reference. They're delimited by markers like:
#
#   <!-- SECTION-CMDS:File Management:BEGIN -->
#   ```bash
#   movie cd <id>
#   movie duplicates
#   ...
#   ```
#   ```powershell
#   movie cd <id>
#   ...
#   ```
#   <!-- SECTION-CMDS:File Management:END -->
#
# Only the bash + powershell fenced blocks between the markers are
# regenerated. The surrounding prose ("Args:", "Assumptions:", "Expected
# output", "If it differs:") is hand-written and preserved verbatim.
SECTION_CMDS_BEGIN = "<!-- SECTION-CMDS:{label}:BEGIN -->"
SECTION_CMDS_END   = "<!-- SECTION-CMDS:{label}:END -->"


def _section_slug(label: str) -> str:
    """
    GitHub-style heading anchor for one of the six command-section labels.

    Mirrors GitHub's slugger:
      - lowercase
      - drop characters that aren't alphanumeric, space, or hyphen
        (importantly: do NOT collapse the surrounding whitespace — the
        characters '&', '_', etc. vanish but the spaces on either side
        remain, so they each become a hyphen)
      - each remaining whitespace character → '-'

    So `&` is dropped (becoming a doubled hyphen between its neighbours):
      "Scanning & Library"      → "scanning--library"
      "Configuration & System"  → "configuration--system"
      "File Management"         → "file-management"
    """
    s = label.lower()
    # Strip non-{alnum, space, hyphen} character-by-character so adjacent
    # spaces are preserved — that's what produces the doubled hyphen in
    # "scanning--library" on real GitHub.
    s = "".join(ch for ch in s if ch.isalnum() or ch in " -")
    # Convert each whitespace character (not runs) to a single hyphen.
    s = "".join("-" if ch == " " else ch for ch in s.strip(" "))
    return s


def _section_anchor(label: str) -> str:
    """'#'-prefixed section anchor for a known command-section label."""
    if label not in SECTION_LABELS:
        sys.exit(f"error: unknown section label '{label}' (not in SECTION_LABELS)")
    return f"#{_section_slug(label)}"


# Pre-compute label → canonical anchor for fast lookup during the rewrite pass.
SECTION_ANCHORS: dict[str, str] = {label: _section_anchor(label) for label in SECTION_LABELS}


# ─── Extra hand-typed doc anchors managed by the rewriter ───────────────────
# These are NOT command sections (no quick-start blocks, no index rows), but
# they're the most-linked top-level README headings and are easy to misspell
# (`#quick_start`, `#QuickStart`, `#installation-guide`, …). Listing them here
# lets _rewrite_section_anchors() fix stale references using the same narrow
# alphanumeric-fingerprint rule it already uses for the six sections.
#
# Add a label here ONLY if README.md actually contains a heading with that
# exact text — the canonical slug is derived via _section_slug(), so a typo
# in this list would silently rewrite real links to a dead anchor.
EXTRA_ANCHOR_LABELS: tuple[str, ...] = (
    "Installation",
    "Quick Start",
    "Troubleshooting",
)

EXTRA_ANCHORS: dict[str, str] = {
    label: f"#{_section_slug(label)}" for label in EXTRA_ANCHOR_LABELS
}


# ─── Custom-anchor whitelist (escape hatch) ─────────────────────────────────
# Anchors listed here are NEVER rewritten and NEVER reported as stale by
# `--check`, even when their fingerprint would otherwise match a managed
# label. Use this for legacy deep-links you intentionally want to keep
# pointing at a custom HTML target (e.g. a hand-authored `<a name="…">`
# anchor that predates the current heading slugs).
#
# Matching is by alphanumeric fingerprint — same rule the rewriter uses —
# so every casing/punctuation variant of an entry is whitelisted together:
#   "legacy-quick-start"  → also covers #legacy_quick_start, #LegacyQuickStart
#
# Two entry shapes are supported:
#   "legacy-link"                            — global: applies anywhere in README
#   ("legacy-link", "Quick Start")           — scoped: applies ONLY when the
#                                              link appears in the document
#                                              section under the named heading
#
# Scoped entries restrict the suppression to one of these section labels:
#   Installation, Quick Start, Troubleshooting
# A scoped entry is a no-op for occurrences of the same anchor in OTHER
# sections — those are still reported and rewritten as normal.
#
# Keep this list short and add a one-line comment per entry explaining why
# the link can't be migrated to the canonical slug.
ANCHOR_WHITELIST: tuple[str | tuple[str, str], ...] = (
    # e.g. "legacy-quick-start",                       # global
    # e.g. ("legacy-quick-start", "Quick Start"),      # only under ## Quick Start
)

# Set of labels that scoped whitelist entries are allowed to reference. Kept
# narrow (just the three EXTRA_ANCHOR_LABELS) so a scoped entry can never be
# attached to a command-section heading where the rewriter is the source of
# truth — those should stay either fully managed or fully whitelisted.
_SCOPABLE_LABELS: frozenset[str] = frozenset(EXTRA_ANCHOR_LABELS)


def _normalize_whitelist_entry(entry: str | tuple[str, str]) -> tuple[str, str | None]:
    """
    Return (anchor, scope_label_or_None). Aborts on malformed entries so a
    typo in ANCHOR_WHITELIST can never silently downgrade a scoped rule
    to a no-op.
    """
    if isinstance(entry, str):
        return (entry, None)
    if (
        isinstance(entry, tuple)
        and len(entry) == 2
        and isinstance(entry[0], str)
        and isinstance(entry[1], str)
    ):
        anchor, scope = entry
        if scope not in _SCOPABLE_LABELS:
            sys.exit(
                f"error: ANCHOR_WHITELIST entry ({anchor!r}, {scope!r}) — "
                f"scope label {scope!r} is not one of "
                f"{sorted(_SCOPABLE_LABELS)}"
            )
        return (anchor, scope)
    sys.exit(
        f"error: ANCHOR_WHITELIST entry has unexpected shape: {entry!r}. "
        "Expected a bare string or a (anchor, section_label) 2-tuple."
    )


# Indexed by fingerprint → set of scopes (None means "global"). A fingerprint
# can appear with multiple scopes; a link is suppressed if its containing
# section matches ANY of them, or if any entry for the fingerprint is global.
_WHITELIST_BY_FINGERPRINT: dict[str, set[str | None]] = {}
for _entry in ANCHOR_WHITELIST:
    _anchor, _scope = _normalize_whitelist_entry(_entry)
    _fp = re.sub(r"[^a-z0-9]", "", _anchor.lower())
    _WHITELIST_BY_FINGERPRINT.setdefault(_fp, set()).add(_scope)

# Backward-compat: code that pre-dates the per-section feature reads this
# frozenset for global suppression. Only entries that have at least one
# global occurrence count — scoped-only entries are NOT in this set so the
# old short-circuit doesn't accidentally apply them everywhere.
_WHITELIST_FINGERPRINTS: frozenset[str] = frozenset(
    fp for fp, scopes in _WHITELIST_BY_FINGERPRINT.items() if None in scopes
)

# Safety net: a whitelist entry that fingerprint-collides with a managed
# label would silently disable that label's auto-fix. That's almost never
# what the maintainer wants — fail loudly at startup instead. We only
# enforce this for *global* entries; a scoped entry is allowed to share a
# fingerprint with a managed label because the scope itself prevents the
# whitelist from short-circuiting outside that one section.
_MANAGED_FINGERPRINTS: frozenset[str] = frozenset(
    re.sub(r"[^a-z0-9]", "", lbl.lower())
    for lbl in (*SECTION_LABELS, *EXTRA_ANCHOR_LABELS)
)
_WHITELIST_COLLISIONS = _WHITELIST_FINGERPRINTS & _MANAGED_FINGERPRINTS
if _WHITELIST_COLLISIONS:
    sys.exit(
        "error: ANCHOR_WHITELIST entries collide with managed labels: "
        f"{sorted(_WHITELIST_COLLISIONS)}. Remove them from the whitelist or "
        "rename the managed label."
    )


# ─── Heading-side ignore list (skip generation + drift reporting) ───────────
# Names listed here are skipped by both the writer and `--check`:
#   - the writer leaves the marker region's body untouched (or, if the
#     markers don't exist yet, doesn't insert anything)
#   - `--check` does NOT report the region as drifted even when its body
#     differs from what the script would otherwise generate
#
# This is the heading/region-side counterpart to ANCHOR_WHITELIST — that one
# silences anchor *targets*, this one silences entire generated *regions*.
# Use it when a section is being deprecated, hand-curated for a release, or
# temporarily frozen while you migrate something.
#
# Region names match exactly what `--check` prints (see _all_regions()):
#   "COMMAND-INDEX:HTML"
#   "COMMAND-INDEX:TEXT"
#   "SECTION-CMDS:Scanning & Library"
#   "SECTION-CMDS:File Management"          ... etc for each SECTION_LABELS entry
#
# Keep this list short and add a one-line comment per entry explaining why
# the region is frozen. Unknown names abort the script — typos can't silently
# leave a real region unmanaged.
IGNORED_REGIONS: tuple[str, ...] = (
    # e.g. "SECTION-CMDS:File Management",  # frozen during v3 migration
)


def _canonical_region_names() -> frozenset[str]:
    """Every region name the script knows how to generate. Used for the typo guard."""
    names = {"COMMAND-INDEX:HTML", "COMMAND-INDEX:TEXT"}
    for label in SECTION_LABELS:
        names.add(f"SECTION-CMDS:{label}")
    return frozenset(names)


_IGNORED_REGION_SET: frozenset[str] = frozenset(IGNORED_REGIONS)
_UNKNOWN_IGNORED = _IGNORED_REGION_SET - _canonical_region_names()
if _UNKNOWN_IGNORED:
    sys.exit(
        "error: IGNORED_REGIONS contains unknown region name(s): "
        f"{sorted(_UNKNOWN_IGNORED)}. Valid names: "
        f"{sorted(_canonical_region_names())}"
    )


def _is_region_ignored(name: str) -> bool:
    """True if `name` is in the IGNORED_REGIONS allowlist."""
    return name in _IGNORED_REGION_SET


def _row_id(command: str) -> str:
    """Stable per-row anchor: 'movie scan <path> --dry-run' → 'movie-scan-path-dry-run'."""
    slug = re.sub(r"[<>]", "", command)
    slug = re.sub(r"\s+--", "-", slug)
    slug = re.sub(r"[\s_]+", "-", slug)
    slug = re.sub(r"-+", "-", slug)
    return slug.strip("-").lower()


def render_html() -> str:
    head = (
        "<table>\n"
        "<thead><tr>\n"
        '<th align="left" width="38%">Command</th>\n'
        '<th align="center" width="4%">→</th>\n'
        '<th align="left" width="20%">Section</th>\n'
        '<th align="left" width="22%">Example keyword</th>\n'
        '<th align="right" width="16%">Anchor</th>\n'
        "</tr></thead>\n"
        "<tbody>"
    )
    body = []
    for command, section, keyword in COMMANDS:
        rid = _row_id(command)
        anchor = SECTION_ANCHORS[section]
        cmd_html = html.escape(command)
        kw_html = html.escape(keyword)
        body.append(
            f'<tr id="{rid}">'
            f'<td><a href="{anchor}" title="Jump to the {section} section"><code>{cmd_html}</code></a></td>'
            f'<td align="center">→</td>'
            f'<td><a href="{anchor}">{section}</a></td>'
            f'<td><code>{kw_html}</code></td>'
            f'<td align="right"><code>#{rid}</code></td>'
            f'</tr>'
        )
    tail = "</tbody>\n</table>"
    return head + "\n" + "\n".join(body) + "\n" + tail


def render_text() -> str:
    cmd_w = max(len(c) for c, _, _ in COMMANDS)
    sec_w = max(len(s) for _, s, _ in COMMANDS)
    rows = ["```text"]
    rows.append(f"{'Command'.ljust(cmd_w)}     {'Section'.ljust(sec_w)}     Anchor")
    anchor_w = max(len(f"#{_row_id(c)}") for c, _, _ in COMMANDS)
    rows.append(f"{'-'*cmd_w}   {'-'*1}   {'-'*sec_w}   {'-'*anchor_w}")
    for command, section, _ in COMMANDS:
        rid = _row_id(command)
        rows.append(f"{command.ljust(cmd_w)}   →   {section.ljust(sec_w)}   #{rid}")
    rows.append("```")
    return "\n".join(rows)


def render_section_block(label: str) -> str:
    """
    Build the bash + powershell quick-start pair for one section.

    Includes every command in `COMMANDS` whose section_label matches, in the
    order they appear in `COMMANDS` (which is alphabetical). The bash and
    powershell blocks are byte-identical because every command name is
    platform-agnostic — the only Windows-specific tip from the previous
    hand-written blocks (quoting `D:\\Media\\...` paths, redirecting with
    `Tee-Object`) is dropped to keep the source-of-truth single.
    """
    cmds = [c for c, sec, _ in COMMANDS if sec == label]
    if not cmds:
        sys.exit(f"error: no commands in section '{label}'")
    body = "\n".join(cmds)
    return f"```bash\n{body}\n```\n```powershell\n{body}\n```"


def _replace_region(content: str, begin: str, end: str, body: str) -> str:
    pattern = re.compile(
        re.escape(begin) + r"\n.*?\n" + re.escape(end),
        flags=re.DOTALL,
    )
    replacement = f"{begin}\n{body}\n{end}"
    new_content, count = pattern.subn(replacement, content)
    if count != 1:
        sys.exit(f"error: expected exactly one '{begin}'..'{end}' region, found {count}")
    return new_content


def _replace_section_blocks(content: str) -> tuple[str, list[str]]:
    """
    Rewrite the bash+powershell pair inside every `<!-- SECTION-CMDS:…:BEGIN
    -->` / `:END -->` region. Returns (new_content, list_of_labels_processed).

    Skips silently if a section's markers don't exist (e.g. Troubleshooting
    has no command list and no markers). Fails loudly if a section IS marked
    but the markers are malformed (begin without end, etc.). Sections whose
    region name appears in IGNORED_REGIONS are left strictly untouched, even
    when their markers are present.
    """
    processed: list[str] = []
    new_content = content
    for label in SECTION_LABELS:
        if _is_region_ignored(f"SECTION-CMDS:{label}"):
            continue  # explicitly frozen — leave the body verbatim
        begin = SECTION_CMDS_BEGIN.format(label=label)
        end = SECTION_CMDS_END.format(label=label)
        if begin not in new_content:
            continue  # section not yet wired up — skip
        if end not in new_content:
            sys.exit(f"error: found '{begin}' without matching '{end}'")
        new_content = _replace_region(new_content, begin, end, render_section_block(label))
        processed.append(label)
    return new_content, processed


def _rewrite_section_anchors(content: str) -> tuple[str, list[str]]:
    """
    Find every Markdown link or HTML href whose anchor target should resolve
    to one of the six command sections, and rewrite the anchor to its current
    canonical slug. Returns (new_content, list_of_human_readable_changes).

    Matching rule (intentionally narrow):
      A `(#xxx)` or `href="#xxx"` is considered "owned by" a section if its
      slug, lowercased and stripped of '-', equals the section label's
      alphanumeric-only fingerprint. So `#file-management`, `#filemanagement`,
      `#FILE-MANAGEMENT` all map to "File Management". The same rule is
      applied to the EXTRA_ANCHOR_LABELS allowlist (Installation, Quick Start,
      Troubleshooting), so `#quick_start`, `#QuickStart`, `#quick-start-guide`
      all collapse to `#quick-start`. Any anchor whose fingerprint matches
      none of the managed labels is left alone. Anchors INSIDE the
      COMMAND-INDEX or per-section quick-start regions are skipped — those
      blocks are fully regenerated by render_html() / render_text() and would
      never contain a stale value at the point this pass runs.
    """
    # label fingerprint (a–z, 0–9 only) → canonical slug (without '#')
    # We keep canonical slug in the value side so the rewriter doesn't have
    # to know whether the match came from SECTION_LABELS or EXTRA_ANCHOR_LABELS.
    fingerprint_to_label: dict[str, str] = {}
    fingerprint_to_canonical: dict[str, str] = {}
    for label in (*SECTION_LABELS, *EXTRA_ANCHOR_LABELS):
        fp = re.sub(r"[^a-z0-9]", "", label.lower())
        # Defensive: if two labels ever collide on fingerprint we'd silently
        # rewrite to whichever wins. Refuse to start instead of corrupting
        # links — the maintainer has to disambiguate the labels.
        if fp in fingerprint_to_label:
            sys.exit(
                f"error: anchor fingerprint collision between "
                f"'{fingerprint_to_label[fp]}' and '{label}' (both → '{fp}')"
            )
        fingerprint_to_label[fp] = label
        fingerprint_to_canonical[fp] = _section_slug(label)

    # Carve out the two index regions so we don't double-process them.
    def _spans(begin: str, end: str) -> tuple[int, int] | None:
        i = content.find(begin)
        j = content.find(end, i + len(begin)) if i != -1 else -1
        return (i, j + len(end)) if i != -1 and j != -1 else None

    skip_spans: list[tuple[int, int]] = []
    for begin, end in ((HTML_BEGIN, HTML_END), (TEXT_BEGIN, TEXT_END)):
        sp = _spans(begin, end)
        if sp:
            skip_spans.append(sp)
    # Per-section quick-start regions are also fully owned by this script —
    # any href inside them is regenerated, so the anchor pass shouldn't
    # double-process them.
    for label in SECTION_LABELS:
        sp = _spans(
            SECTION_CMDS_BEGIN.format(label=label),
            SECTION_CMDS_END.format(label=label),
        )
        if sp:
            skip_spans.append(sp)

    def _in_skip(pos: int) -> bool:
        return any(s <= pos < e for s, e in skip_spans)

    # Per-section whitelist support: pre-compute the H2 section span for each
    # _SCOPABLE_LABELS entry so we can answer "what scoped section contains
    # position N?" in O(1). Sections that don't appear in this README are
    # simply absent from the map — a scoped whitelist entry pointing at one
    # of them then never matches anything, which is the safe outcome.
    scope_spans: dict[str, tuple[int, int]] = {}
    if _WHITELIST_BY_FINGERPRINT:
        # Only bother scanning the document when at least one whitelist
        # entry exists — keeps the unmodified-default path zero-cost.
        h2_pat = re.compile(r"^## +(.+?)\s*$", flags=re.MULTILINE)
        h2_matches = list(h2_pat.finditer(content))
        for idx, m in enumerate(h2_matches):
            heading_text = m.group(1).strip()
            # Strip leading emoji + whitespace before comparing — the
            # README uses headings like "## ✨ Highlights" but "## Quick Start"
            # for the scopable labels. Belt and braces: also try the raw text.
            stripped = re.sub(r"^[^A-Za-z]+", "", heading_text).strip()
            for label in _SCOPABLE_LABELS:
                if label in (heading_text, stripped):
                    end = h2_matches[idx + 1].start() if idx + 1 < len(h2_matches) else len(content)
                    # First-occurrence wins; later duplicates are ignored so a
                    # single label never has ambiguous spans.
                    scope_spans.setdefault(label, (m.start(), end))

    def _scope_at(pos: int) -> str | None:
        """Return the scopable section label whose span contains `pos`, or None."""
        for label, (s, e) in scope_spans.items():
            if s <= pos < e:
                return label
        return None

    def _is_whitelisted(fp: str, pos: int) -> bool:
        scopes = _WHITELIST_BY_FINGERPRINT.get(fp)
        if not scopes:
            return False
        if None in scopes:
            return True  # at least one global entry covers this fingerprint
        # All entries for this fingerprint are scoped — only suppress when
        # the link is physically inside one of those sections.
        return _scope_at(pos) in scopes

    changes: list[str] = []

    # Patterns: Markdown link target `(#xxx)` and HTML attribute `href="#xxx"`.
    md_pat = re.compile(r"\(#([A-Za-z0-9_-]+)\)")
    html_pat = re.compile(r'href="#([A-Za-z0-9_-]+)"')

    def _maybe_rewrite(match: re.Match[str], wrap: tuple[str, str]) -> str:
        if _in_skip(match.start()):
            return match.group(0)
        raw = match.group(1)
        fp = re.sub(r"[^a-z0-9]", "", raw.lower())
        if _is_whitelisted(fp, match.start()):
            return match.group(0)  # explicitly whitelisted — leave verbatim
        label = fingerprint_to_label.get(fp)
        if label is None:
            return match.group(0)  # not a known section — leave alone
        canonical = fingerprint_to_canonical[fp]
        if raw == canonical:
            return match.group(0)
        line_no = content.count("\n", 0, match.start()) + 1
        changes.append(f"  line {line_no}: #{raw} → #{canonical}  ({label})")
        return f"{wrap[0]}#{canonical}{wrap[1]}"

    new_content = md_pat.sub(lambda m: _maybe_rewrite(m, ("(", ")")), content)
    new_content = html_pat.sub(lambda m: _maybe_rewrite(m, ('href="', '"')), new_content)
    return new_content, changes


# ─── Rich --check diagnostics ───────────────────────────────────────────────
# These helpers re-derive what changed (anchor lines + which generated region
# drifted) so `--check` can show actionable context, not just "stale".
# They never mutate state; they're only read from the --check failure branch.

# An auto-generated region this script owns. `name` is the human label printed
# in error output; `begin`/`end` are the marker comments wrapping the block.
_Region = tuple[str, str, str]  # (name, begin_marker, end_marker)


def _all_regions() -> list[_Region]:
    """Every begin/end marker pair this script regenerates, in scan order."""
    regions: list[_Region] = [
        ("COMMAND-INDEX:HTML", HTML_BEGIN, HTML_END),
        ("COMMAND-INDEX:TEXT", TEXT_BEGIN, TEXT_END),
    ]
    for label in SECTION_LABELS:
        regions.append((
            f"SECTION-CMDS:{label}",
            SECTION_CMDS_BEGIN.format(label=label),
            SECTION_CMDS_END.format(label=label),
        ))
    return regions


def _extract_region_body(content: str, begin: str, end: str) -> str | None:
    """Return the text strictly between `begin\\n` and `\\n{end}`, or None if absent."""
    i = content.find(begin)
    if i == -1:
        return None
    body_start = i + len(begin) + 1  # skip the trailing '\n' of the BEGIN line
    j = content.find(end, body_start)
    if j == -1:
        return None
    return content[body_start:j - 1]  # strip the leading '\n' before END


def _line_at(content: str, line_no: int) -> str:
    """1-indexed line lookup, returning '' for out-of-range to keep callers simple."""
    lines = content.splitlines()
    if 1 <= line_no <= len(lines):
        return lines[line_no - 1]
    return ""


def _nearest_heading_above(content: str, line_no: int) -> tuple[int, str] | None:
    """
    Return (heading_line_no, heading_text) for the closest Markdown heading
    AT ANY LEVEL (H1–H6) at or above `line_no`. Returns None if no heading
    precedes the position. Heading text is returned verbatim, including the
    leading '#' run, so callers can render it as-is for visual context.

    Robust against ATX look-alikes that aren't real headings:
      * Fenced code blocks using ``` or ~~~. Per CommonMark, a fence opens
        with ≥3 of the same char (indented up to 3 spaces) and is closed
        only by a fence using the SAME char with length ≥ the opener's.
        Inner fences with a different char or shorter length do NOT toggle
        state — this is what "nested fenced code block" robustness means.
      * Indented code blocks (lines starting with ≥4 spaces or a tab),
        which CommonMark treats as code. A `#` here is not a heading.
      * ATX headings themselves may be indented up to 3 spaces; 4+ spaces
        of leading whitespace disqualifies them.

    Also recognises setext headings — a non-blank text line followed by an
    underline of '=' (H1) or '-' (H2). The reported line number is the
    text line (so the breadcrumb points at what the user reads as the
    heading), and the reported text is the text line verbatim.
    """
    return _nearest_heading_above_levels(content, line_no, _ALL_HEADING_LEVELS)


# Default set used when no caller-specific filter is provided. Frozen so it
# can safely be passed around as a default argument without aliasing risk.
_ALL_HEADING_LEVELS: frozenset[int] = frozenset(range(1, 7))


def _nearest_heading_above_levels(
    content: str,
    line_no: int,
    levels: frozenset[int],
) -> tuple[int, str] | None:
    """
    Same as _nearest_heading_above but restricted to a caller-supplied set
    of heading levels (1–6). A heading whose level is NOT in `levels` is
    skipped entirely — the walk continues looking for an older heading at
    a permitted level. Returns None if none qualify.

    Used by `--check` so reviewers can scope the breadcrumb to, e.g., only
    H2 sections (matching the per-section whitelist's scoping model).

    Setext headings count as level 1 ('===' underline) or level 2
    ('---' underline) for the purposes of `levels` filtering.
    """
    lines = content.splitlines()
    if line_no < 1 or not lines:
        return None
    # Active fence descriptor: (char, min_close_length) or None.
    fence: tuple[str, int] | None = None
    found: tuple[int, str] | None = None
    # Setext-heading state: the most recent line that could serve as the
    # text line of a setext heading. Cleared on blanks, ATX headings,
    # fence opens, indented-code lines, or any line consumed by a fence.
    # Stored as (line_no_1indexed, text_verbatim). None when no candidate.
    prev_para: tuple[int, str] | None = None
    # Walk forward up to the target line. Forward-walk (not backward)
    # because fence state only resolves correctly in document order.
    upper = min(line_no, len(lines))
    for idx in range(upper):
        line = lines[idx]
        # Count leading spaces (tabs count as code-block indent).
        indent = len(line) - len(line.lstrip(" "))
        starts_with_tab = line.startswith("\t")
        stripped = line.lstrip(" ")
        # --- Fence handling ---
        if fence is not None:
            # Look for a closing fence of the same char, length ≥ opener,
            # with ≤3 spaces of indent and only whitespace after the run.
            ch, min_len = fence
            if indent <= 3 and stripped.startswith(ch * min_len):
                run = len(stripped) - len(stripped.lstrip(ch))
                tail = stripped[run:]
                if run >= min_len and tail.strip() == "":
                    fence = None
            # Either way, lines inside a fence cannot be headings.
            prev_para = None
            continue
        # Opening fence: ≤3 spaces indent, ≥3 of ` or ~. The info string
        # may contain anything except backticks (for ``` fences).
        if indent <= 3 and (stripped.startswith("```") or stripped.startswith("~~~")):
            ch = stripped[0]
            run = len(stripped) - len(stripped.lstrip(ch))
            info = stripped[run:]
            # ``` fences forbid backticks in the info string; ~~~ has no
            # such restriction. A bad info string means it's not a fence.
            if ch == "`" and "`" in info:
                pass  # not a real fence opener
            else:
                fence = (ch, run)
                prev_para = None
                continue
        # Indented code block: 4+ leading spaces or a leading tab. ATX
        # headings are disallowed here per CommonMark.
        if indent >= 4 or starts_with_tab:
            prev_para = None
            continue
        # Blank line — terminates any setext-text candidate. CommonMark
        # requires the underline to immediately follow a non-blank line.
        if stripped.strip() == "":
            prev_para = None
            continue
        # --- Setext underline detection ---
        # Underline line: ≤3 spaces indent, all chars are '=' or all '-'
        # (only one kind, ≥1 char), optionally followed by trailing spaces.
        # Promotes the immediately-preceding paragraph text line into a
        # setext heading. '-'-only underline with no preceding paragraph
        # text is a thematic break — left to fall through as content.
        setext = _setext_underline_level(stripped) if prev_para is not None else None
        if setext is not None:
            level = setext  # 1 for '=', 2 for '-'
            if level in levels:
                # Report at the text line (what the user sees as the heading).
                found = (prev_para[0], prev_para[1].rstrip())
            # The underline line is consumed; it is not itself paragraph text.
            prev_para = None
            continue
        # ATX heading: optional 1–3 spaces, then 1–6 '#', then a space and
        # at least one non-space char. Trailing '#' run is allowed but we
        # keep the line verbatim for display.
        m = re.match(r"^ {0,3}(#{1,6}) +(\S.*)$", line)
        if m and len(m.group(1)) in levels:
            found = (idx + 1, line.rstrip())
            prev_para = None
            continue
        if m:
            # ATX heading at a disallowed level — still consumed, not paragraph text.
            prev_para = None
            continue
        # Otherwise: a non-blank, non-heading, non-code line. It becomes
        # the candidate text for any setext underline that follows.
        prev_para = (idx + 1, line)
    return found


def _setext_underline_level(stripped: str) -> int | None:
    """
    Return 1 if `stripped` is a setext H1 underline ('=' run, optional
    trailing spaces), 2 if it's a setext H2 underline ('-' run, ditto),
    or None otherwise.

    Caller is responsible for ensuring leading indent is ≤3 (already
    handled by stripping leading spaces and checking before calling).
    """
    if not stripped:
        return None
    head = stripped[0]
    if head not in ("=", "-"):
        return None
    run = len(stripped) - len(stripped.lstrip(head))
    tail = stripped[run:]
    # Underline body must be ≥1 char and trailing must be only whitespace.
    if run < 1 or tail.strip() != "":
        return None
    return 1 if head == "=" else 2


# Default breadcrumb level set when --breadcrumb-levels is not passed.
# Frozen so it can safely be reused as a default across calls.
_DEFAULT_BREADCRUMB_LEVELS: frozenset[int] = frozenset({2})


def _parse_breadcrumb_levels(raw: str) -> frozenset[int]:
    """
    Parse the --breadcrumb-levels CSV value into a frozenset[int] of
    permitted heading levels (1-6).

    Behaviour:
      * Empty tokens (from "1,,2") are silently ignored.
      * Non-integer tokens emit a stderr warning and are dropped.
      * Integers outside 1-6 are clamped to the nearest endpoint and a
        stderr warning names the original value (so the user can see
        what was clamped, rather than silently morphing 0 into 1).
      * Duplicates collapse via the set.
      * If the result is empty after parsing, falls back to H2 with a
        warning so --check always produces some breadcrumb.
    """
    levels: set[int] = set()
    for token in raw.split(","):
        token = token.strip()
        if not token:
            continue
        try:
            value = int(token)
        except ValueError:
            sys.stderr.write(
                f"warning: --breadcrumb-levels: ignoring non-integer "
                f"token {token!r}\n"
            )
            continue
        if value < 1:
            sys.stderr.write(
                f"warning: --breadcrumb-levels: clamping {value} to 1\n"
            )
            value = 1
        elif value > 6:
            sys.stderr.write(
                f"warning: --breadcrumb-levels: clamping {value} to 6\n"
            )
            value = 6
        levels.add(value)
    if not levels:
        sys.stderr.write(
            "warning: --breadcrumb-levels: no valid levels parsed, "
            "falling back to H2 only\n"
        )
        return _DEFAULT_BREADCRUMB_LEVELS
    return frozenset(levels)


def _format_anchor_change(
    content: str,
    change: str,
    breadcrumb_levels: frozenset[int] = _DEFAULT_BREADCRUMB_LEVELS,
) -> str:
    """
    Decorate one anchor-change record with its source line and a region tag.

    Input format (produced by _rewrite_section_anchors):
        '  line 42: #FOO → #foo  (Foo)'
    Output:
        '  line 42 [outside index]: #FOO → #foo  (Foo)\\n'
        '      under: ## ✨ Quick Start (line 70)\\n'
        '      > <line text>'

    `breadcrumb_levels` controls which heading levels (1–6) qualify for
    the `under:` line. Defaults to H2 only, matching the per-section
    whitelist's scoping model. Pass `_ALL_HEADING_LEVELS` to restore the
    nearest-heading-at-any-level behaviour.
    """
    m = re.match(r"\s*line (\d+):", change)
    if not m:
        # Should never happen, but degrade gracefully rather than crash --check.
        return change
    line_no = int(m.group(1))
    src = _line_at(content, line_no).rstrip()
    # The rewriter already excludes anchors inside generated regions, so every
    # reported change is by construction outside them — surface that explicitly
    # so reviewers know where to edit.
    tagged = re.sub(r"line (\d+):", r"line \1 [outside index]:", change, count=1)
    # Nearest heading at one of the configured levels — gives reviewers a
    # one-line breadcrumb to the section that owns the offending link.
    heading = _nearest_heading_above_levels(content, line_no, breadcrumb_levels)
    if heading is not None:
        h_line, h_text = heading
        tagged += f"\n      under: {h_text} (line {h_line})"
    if src:
        tagged += f"\n      > {src}"
    return tagged


def _stale_regions(original: str, updated: str) -> list[tuple[str, str]]:
    """
    Compare each owned region in `original` vs `updated`. Return a list of
    (region_name, unified_diff_text) for every region whose body changed.
    Regions missing from `original` (e.g. Troubleshooting has no markers) are
    silently skipped — they're not drift, they're just not wired up. Regions
    listed in IGNORED_REGIONS are also skipped, even when their body differs
    from what the script would generate.
    """
    drifted: list[tuple[str, str]] = []
    for name, begin, end in _all_regions():
        if _is_region_ignored(name):
            continue  # explicitly frozen — drift is allowed
        before = _extract_region_body(original, begin, end)
        after = _extract_region_body(updated, begin, end)
        if before is None or after is None:
            continue  # region not present — not stale, just absent
        if before == after:
            continue
        diff = "\n".join(
            difflib.unified_diff(
                before.splitlines(),
                after.splitlines(),
                fromfile=f"{name} (current)",
                tofile=f"{name} (expected)",
                n=3,
                lineterm="",
            )
        )
        drifted.append((name, diff))
    return drifted


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="exit 1 if README would change")
    parser.add_argument(
        "--list-sections",
        action="store_true",
        help="print the canonical 'label -> #slug' map for the six sections and exit",
    )
    parser.add_argument(
        "--breadcrumb-levels",
        default="2",
        metavar="LEVELS",
        help=(
            "comma-separated heading levels (1-6) that count for the "
            "'under:' breadcrumb in --check output. Default: 2 (H2 only). "
            "Out-of-range or non-integer tokens are clamped/ignored with "
            "a warning; if nothing valid remains, falls back to H2."
        ),
    )
    args = parser.parse_args()
    breadcrumb_levels = _parse_breadcrumb_levels(args.breadcrumb_levels)

    if args.list_sections:
        all_labels = (*SECTION_LABELS, *EXTRA_ANCHOR_LABELS)
        # Pre-normalize whitelist entries so width math and rendering work
        # for both the bare-string and (anchor, scope) tuple forms.
        whitelist_normalized = [_normalize_whitelist_entry(e) for e in ANCHOR_WHITELIST]
        width = max((len(label) for label in all_labels), default=0)
        width = max(width, *(len(anchor) for anchor, _ in whitelist_normalized), 0)
        print("# command sections")
        for label in SECTION_LABELS:
            print(f"{label.ljust(width)}  ->  {SECTION_ANCHORS[label]}")
        print("# extra doc anchors (auto-fix only)")
        for label in EXTRA_ANCHOR_LABELS:
            print(f"{label.ljust(width)}  ->  {EXTRA_ANCHORS[label]}")
        print("# whitelist (never rewritten, never reported)")
        if not whitelist_normalized:
            print("(empty)")
        for anchor, scope in whitelist_normalized:
            tag = "[verbatim]" if scope is None else f"[scoped to ## {scope}]"
            print(f"{anchor.ljust(width)}  ->  #{anchor}  {tag}")
        print("# ignored regions (never regenerated, drift suppressed)")
        if not IGNORED_REGIONS:
            print("(empty)")
        for region in IGNORED_REGIONS:
            print(f"{region}  [frozen]")
        return 0

    original = README.read_text(encoding="utf-8")

    # 1. Rewrite stale section anchors first so the COMMAND-INDEX regions are
    #    rebuilt from the same canonical slugs everything else now uses.
    updated, anchor_changes = _rewrite_section_anchors(original)

    # 2. Regenerate the two index regions in place — unless they're frozen
    #    via IGNORED_REGIONS, in which case we leave the body verbatim and
    #    drift is suppressed in the --check path.
    if not _is_region_ignored("COMMAND-INDEX:HTML"):
        updated = _replace_region(updated, HTML_BEGIN, HTML_END, render_html())
    if not _is_region_ignored("COMMAND-INDEX:TEXT"):
        updated = _replace_region(updated, TEXT_BEGIN, TEXT_END, render_text())

    # 3. Regenerate the per-section quick-start bash+powershell pairs.
    updated, sections_done = _replace_section_blocks(updated)

    if args.check:
        if updated != original:
            sys.stderr.write("README.md is stale:\n")

            # 1. Anchor rewrites — show the offending source line for each.
            if anchor_changes:
                sys.stderr.write(
                    f"\n  {len(anchor_changes)} section anchor(s) need "
                    "rewriting:\n"
                )
                for change in anchor_changes:
                    sys.stderr.write(
                        _format_anchor_change(original, change, breadcrumb_levels)
                        + "\n"
                    )

            # 2. Generated-region drift — name each stale region and emit a
            #    unified diff so the failure is reviewable straight from CI logs.
            drifted = _stale_regions(original, updated)
            if drifted:
                sys.stderr.write(
                    f"\n  {len(drifted)} generated region(s) need regenerating:\n"
                )
                for name, diff in drifted:
                    sys.stderr.write(f"    - {name}\n")
                    if diff:
                        # Indent the diff so it's visually nested under the region name.
                        for line in diff.splitlines():
                            sys.stderr.write(f"      {line}\n")

            # 3. Belt-and-braces: if neither bucket caught the drift, say so
            #    explicitly rather than printing a confusingly-empty failure.
            if not anchor_changes and not drifted:
                sys.stderr.write(
                    "  README would change but no specific region or anchor "
                    "could be attributed — run the script to inspect the diff.\n"
                )

            sys.stderr.write("\nRun: python3 scripts/gen-command-index.py\n")
            return 1
        msg = (
            "README.md command index, section anchors, and per-section "
            f"quick-start blocks ({len(sections_done)} sections) are up to date."
        )
        if IGNORED_REGIONS:
            msg += f" Ignored regions: {len(IGNORED_REGIONS)} ({', '.join(IGNORED_REGIONS)})."
        print(msg)
        return 0

    if updated == original:
        msg = (
            "README.md command index, section anchors, and per-section "
            f"quick-start blocks ({len(sections_done)} sections) already up to date."
        )
        if IGNORED_REGIONS:
            msg += f" Ignored regions: {len(IGNORED_REGIONS)} ({', '.join(IGNORED_REGIONS)})."
        print(msg)
        return 0
    README.write_text(updated, encoding="utf-8")
    msg = (
        f"README.md command index regenerated ({len(COMMANDS)} commands), "
        f"{len(sections_done)} per-section quick-start block(s) refreshed."
    )
    if anchor_changes:
        msg += f"\nRewrote {len(anchor_changes)} section anchor reference(s):"
        for change in anchor_changes:
            msg += "\n" + change
    print(msg)
    return 0


if __name__ == "__main__":
    sys.exit(main())