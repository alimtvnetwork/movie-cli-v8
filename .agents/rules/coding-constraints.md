# Essential Coding Constraints

> **Scope:** Repository-wide (Go, Python, TypeScript, Documentation)  
> **Source:** `AGENTS.md` Section 8 & `02-spec/02-coding-guidelines/`

---

## 1. Lowercase Naming Convention

- All files, scripts, documentation, and system files generated or modified by the AI MUST use strictly lowercase naming (e.g., `readme.md`, `agents.md`, `skill.md`).
- Zero exceptions for uppercase letters in file names.

## 2. Relative Git Paths Mandate

- All paths, markdown links, and subtask references MUST be relative paths starting from the git repository root.
- Absolute filesystem paths (`file:///`, `/absolute/...`, `C:\...`) are strictly prohibited in repository files and markdown.

## 3. Parameter Structs

- Loose functions with more than 2-3 parameters are banned.
- Use explicit `*Params` structs to bundle parameters.

## 4. Vertical Line Gaps

- Mandatory blank lines:
  - Before every `if` block (unless at start of block).
  - After every closing `}` (unless followed by `}`, `else`, `case`, or `catch`).
  - Before every `return` statement (unless only statement in block).
  - Around multiline struct instantiations and call chains.
- Never two consecutive blank lines anywhere.

## 5. Micro-Batching

- Refactors and task executions must be broken into bounded batches of 5-8 files.
- Never attempt broad repo-wide rewrites in a single unbounded step.
