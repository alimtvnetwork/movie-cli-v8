---
name: strictly-prohibited
description: Hard prohibitions — never write time/date content in readme.txt, never re-suggest rejected items, never propose git-update-time anywhere in the README
type: constraint
---

# Strictly Prohibited (mirror of spec/00-spec-authoring-guide/10-strictly-prohibited.md)

These rules are absolute. Treat each as a hard constraint in every session.
Never violate, never re-suggest, never propose workarounds.

## P-001 — readme.txt time content
- Never write, generate, suggest, or include ANY time-related content in
  `readme.txt`: no date, no time, no timestamp, no "generated at",
  no "last updated", no Malaysia time, no UTC, no 12h/24h clock,
  no ISO date, no "git update time", no "last commit time", no "sync
  time", and no equivalent under a different name.
- This applies to the file content AND to any suggestion/proposal about
  the file in chat, commits, plans, or future enhancements.

## P-002 — Re-proposing prohibited items
- Never reintroduce a prohibited item under a different framing,
  synonym, or "optional" workaround.
- When asked to do something prohibited, refuse and cite the P-### id.

**Why:** Project owner explicitly rejected time content in readme.txt.
Re-suggesting wastes their time and breaks the prohibition contract.
