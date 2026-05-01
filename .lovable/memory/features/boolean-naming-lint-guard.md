---
name: Boolean naming lint guard
description: v2.307.0 CI guard scripts/check-boolean-naming.sh enforces mem://constraints/boolean-no-negative-words by failing the build if any Go identifier matches (Is|Has)(Un|Not|No)<letter>; allow-lists stdlib os.IsNotExist plus common English words (IsUnique, IsNoteworthy, HasNotes, IsUntitled, IsNormal, IsUnion, …); wired into ci.yml lint job
type: feature
---
**Rule (now CI-enforced)**: Boolean identifiers must use positive
semantic synonyms, never `Un` / `Not` / `No` prefixes after `Is` /
`Has`. Examples:
- `IsUndone` → `IsReverted`
- `IsNotReady` → `IsPending`
- `HasNoChild` → `IsLeaf`
- `IsUnsaved` → `IsDirty`

**Script**: `scripts/check-boolean-naming.sh [root=.]`
- Pattern: `(Is|Has)(Un|Not|No)[A-Za-z]` — catches both lowercase tail
  (`IsUndone`) and uppercase tail (`HasNoChild`).
- Strips `// …` line comments before matching so warning banners
  don't false-positive.
- Allow-lists stdlib `os.IsNotExist` family.
- Allow-lists English words that legitimately start with Un/Not/No:
  `IsUnique`, `IsUniversal`, `IsUnion`, `IsUnit`, `IsUniform`,
  `IsUnknown`, `IsUntitled`, `IsUnread`, `IsNoteworthy`, `IsNotable`,
  `HasNotes`, `IsNormal`, `IsNominal`, etc.
- Verified: catches `IsUndone`, `HasNoChild`, `IsUnsaved`, `IsNotReady`;
  ignores `IsUnique`, `IsNoteworthy`, `HasNotes`, `IsUntitled`,
  `IsNormal`; passes on current repo.

**Companion to**: `scripts/check-lastinsertid-anti-pattern.sh`
(v2.306.0). Both wired into `.github/workflows/ci.yml` lint job
right after the Forbidden term guard.
