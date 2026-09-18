# Boolean Principles (Cross-Language)

> **Scope:** Repository-wide (Go, TypeScript, Python, Shell, Specs)  
> **Source:** `02-spec/02-coding-guidelines/01-cross-language/02-boolean-principles/01-index.md`

---

## 1. Positive Framing Mandate

- All boolean identifiers (variables, struct fields, properties, function names) MUST be framed positively.
- Permitted prefixes are `is` and `has` ONLY (`Is` / `Has` for PascalCase; e.g. `isValid`, `hasData`, `isReady`, `hasPermission`).
- Prefixes such as `can`, `should`, `was`, `will`, `did`, `must`, and negative prefixes (`not`, `non`, `no`) are strictly banned.

## 2. No Explicit True Checks (TOTAL BAN)

- NEVER evaluate a boolean explicitly against `true` or `false`:
  - ❌ `if isReady == true`
  - ❌ `if (isValid === true)`
  - ❌ `if hasAccess == false`
- Positive booleans MUST ALWAYS be evaluated implicitly:
  - ✅ `if isReady { ... }`
  - ✅ `if isValid { ... }`

## 3. No Mixed Polarity

- NEVER combine a positive check and a negative check in the same `if` condition:
  - ❌ `if isA && !isB`
  - ❌ `if (hasToken && !isExpired)`
- Invert or extract to a positive named boolean before branching:
  - ✅ `if isA && isBDisabled` or extract `isExecutable := isA && !isB` followed by `if isExecutable`

## 4. No Inverted Success Checks

- NEVER check `!response.isSuccess` or `!result.isOk`.
- Explicitly evaluate failure models or positive failure flags:
  - ✅ `if response.isFail { ... }`
  - ✅ `if appErr != nil { ... }`

## 5. Maximum Two Operands & Single Operator

- Avoid complex chained joins:
  - ❌ `if isA && isB && isC`
  - ❌ `if isA || (isB && isC)`
- Extract combined predicates into dedicated boolean functions or intermediate variables.
