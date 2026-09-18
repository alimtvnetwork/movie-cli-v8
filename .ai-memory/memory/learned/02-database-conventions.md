# Learned: Database Conventions

- **Tables**: Singular PascalCase (e.g. `User`).
- **PKs**: `{TableName}Id` (`INTEGER PRIMARY KEY AUTOINCREMENT`). No UUIDs unless required.
- **FKs**: Matches PK name exactly.
- **Booleans**: `Is`/`Has` prefix, positive-only (`IsActive`).
- **Engine**: SQLite Split DB architecture is the default. MySQL is fallback. No raw SQL.
