# Boolean Naming and Truthiness Standard

## Core Principles
1. **Positive Logic Only:** Always use isDefined or positive checks. NEVER use inverted negative checks like !isEmpty or !is_empty.
2. **Boolean Prefixes:** Boolean variables and methods MUST be prefixed with is, has, can, etc. (e.g., isReady, hasData).
3. **Implicit Checks:** NEVER explicitly check against 	rue or alse (e.g., if isReady == true is a TOTAL BAN).
4. **Map Lookups:** Use positive existence checks: if val, isDefined := m[key]; isDefined { ... }.
