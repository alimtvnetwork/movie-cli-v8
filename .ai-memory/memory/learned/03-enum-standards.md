# Learned: Enum Standards

- **Source of Truth**: Defined in YAML manifests under `02-spec/<module>/enums/`.
- **Generation**: Generated using `scripts/codegen/gen-all-enums.mjs`. Never hand-edited.
- **Naming**: PascalCase. PHP enums must end with `Type`.
- **Parsing**: Always use generated parsing methods instead of raw strings.
