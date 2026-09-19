# Database Conventions Enforcements

1. **Table Names**: Singular PascalCase (e.g. `User`, `Project`). Never plural.
2. **Column Names**: PascalCase (e.g. `PluginSlug`, `CreatedAt`).
3. **Primary Keys**: `{TableName}Id` (e.g. `UserId`). Must be `INTEGER PRIMARY KEY AUTOINCREMENT`. No UUIDs unless required.
4. **Foreign Keys**: Exact PK name from referenced table.
5. **Booleans**: `Is`/`Has` prefix, positive-only (`IsActive`).
6. **No Raw SQL**: Use ORMs. Create views for joins rather than raw SQL joins in code.
7. **Default Engine**: SQLite (Split DB pattern). MySQL is fallback.
