# Enum Standards Standard

## Core Principles
1. **Enum Suffix:** All enums must end with Type (e.g., LogLevelType, ProcessStateType).
2. **Value Suffix:** Enum constant values must suffix the enum name (e.g., InfoLogLevelType, ErrorLogLevelType).
3. **Min/Max Boundaries:** Enums must define Min() and Max() boundaries for validation.
4. **BaseEnumer:** Go enums must embed or utilize BaseEnumer for parsing and validation.
5. **No Magic Strings:** Never use raw strings in place of enums.
