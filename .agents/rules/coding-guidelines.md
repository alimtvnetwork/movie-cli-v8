# Coding Guidelines Enforcements

1. **No Generated Code or Artifacts**: Never commit generated code, cache files, test outputs, or binaries. Keep .gitignore updated.
2. **Function Size Caps**: Functions must be 8 lines preferred, 15 lines max. Files max 300 lines (React 100).
3. **No Nested If**: Zero nesting. Use early returns and guard clauses.
4. **Positive Conditions**: `if` conditions must be positive. No `!` operators on complex checks.
5. **No Swallowed Errors**: Catch blocks must log operation name + key inputs and rethrow/handle. No empty catches.
6. **Narrow Types**: No `any`, `unknown`, `interface{}`, `object`, `dynamic`. Narrow types immediately at trust boundaries.
7. **No Magic Strings/Numbers**: Use typed constants or enums.
8. **Definitions Location**: Types, enums, constants get their own files.
9. **DRY Rule**: Duplicate logic across two sites means extract it immediately.
10. **Immutable First**: Assign variables once. Use `const`/`let` instead of `let mut` or `var`. Build result objects with copy/spread, not mutation.
