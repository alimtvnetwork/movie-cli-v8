# Database Conventions Standard

## Core Principles
1. **Singular Tables:** All table names must be singular and PascalCase (e.g., User, Project).
2. **PascalCase Columns:** All columns and fields must be PascalCase (e.g., CreatedAt, PluginSlug).
3. **Primary Keys:** The primary key must always be {TableName}Id (e.g., UserId, ProjectId) and type INTEGER PRIMARY KEY AUTOINCREMENT.
4. **Foreign Keys:** Must exactly match the primary key name of the referenced table.
5. **No UUIDs:** Use INTEGER instead of UUIDs unless strictly required.
