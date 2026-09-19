# Enum Standards Enforcements

1. **Source of Truth**: All cross-language enums must be defined in `02-spec/<module>/enums/<EnumName>.yaml`.
2. **Code Generation**: Use `scripts/codegen/gen-all-enums.mjs` to generate enums across Go, TS, PHP, and Rust.
3. **No Hand-Editing**: Never manually edit generated enum files (`*_generated.go`, etc.).
4. **Naming**: PascalCase for enum names and variants. PHP requires `Type` suffix.
5. **No Magic Strings**: Use generated parse methods (`ParseEnum`, `from`, `parse`) instead of raw strings or numbers.
