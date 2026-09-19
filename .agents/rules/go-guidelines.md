# Rule: Go Language Guidelines

1. **No Underscores:** Go identifiers, structs, fields, and variables must use `camelCase` or `PascalCase`. Snake_case is strictly banned.
2. **Error Types:** Functions returning errors must return `*appfault.AppError`.
3. **Parameter Structs:** Avoid functions with more than 2-3 parameters; encapsulate into `*Params` structs.
4. **Single Result Container Return Types:** Functions returning errors paired with collections or values must return `appfault.ResultMap[K, V]`, `appfault.ResultSlice[T]`, or `appfault.Result[T]`, never bare error tuples. Side-effects only return `*appfault.AppError`.
5. **Pointer-Attached Null Safety & Method Composition:** Result inspection methods MUST attach to pointer receivers (`(r *Result[T])`) with line-1 `if r == nil` guards returning safe defaults. Methods MUST compose existing methods (`r.IsFailure()`, `r.IsSuccess()`, `r.Count()`). Callers use `IsCountOtherThan(N)`, `IsEmpty()`, `HasRecord()`, `IsDefined()`.
6. **Affirmative Boolean Parameters & Fields:** Never use single-letter boolean parameters (`v bool`, `b bool`) or bare verbs/nouns (`stop bool`, `pause bool`, `defined bool`). Always use affirmative names: `isStopOnFail bool`, `isStopped bool`, `isPaused bool`, `isDefined bool`.
7. **Types Folder & types.go Single Type Reuse:** All domain payload structs (e.g. `ScheduleExportBundle`) and repeated generic Result envelopes (`type ScheduleExportBundleResult = result.ResultSlice[ScheduleExportBundle]`) MUST be defined in a dedicated `types.go` file within each package as a single reusable named type. Never declare unexported structs or raw generic Result envelopes inline in implementation files.
