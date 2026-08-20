# Database design specifications

Naming, columns, and indexes. When you add or change a table, the executable SQL in the repo wins — do not copy only the examples in this page.

## Aligned with this repo

- Scripts: `scripts/db/{mysql,sqlite,oracle}/`, entry `init.sql` (relative `.read` / `source`)
- Default database: `sqlite_main` in `configs/database.yaml`, file `./scripts/data/gateway.db`. MySQL / Oracle schema name is `gateway`, not the retired `web_hub_here`
- Table names: `HUB_` prefix, uppercase, underscores (`HUB_GW_INSTANCE`, `HUB_USER`)
- Columns: camelCase; no autoincrement; never name a column `id`
- Primary key: most business tables use `(tenantId, <businessId>)`; check that table’s `.sql`
- Booleans: `activeFlag` / `statusFlag` with `'Y'` / `'N'`, not `is*` prefixes
- Reserved columns: the convention is `reserved1`–`reserved10`; some tables (e.g. `HUB_GW_INSTANCE`) only have `reserved1`–`reserved5`. Follow the script

---

## Database name

- Lowercase, words separated by underscores
- SQLite: a file path, e.g. `./scripts/data/gateway.db`
- MySQL / Oracle: schema or service name, e.g. `gateway`

---

## Table names

- Start with `HUB`, uppercase, underscores
- Module + function, singular: `HUB_USER`, `HUB_GW_INSTANCE`
- Avoid reserved SQL keywords

---

## Column names

- camelCase
- Do not use `id` as a column name; use `userId`, `gatewayInstanceId`, …
- No autoincrement
- Foreign keys should name the related entity: `parentUserId`

**Flags:** `activeFlag`, `deleteFlag` — not `isDeleted`, `isActive`.

**Datetimes:** `DATETIME`. `addTime` defaults to `CURRENT_TIMESTAMP`. SQLite scripts usually do not use `ON UPDATE` for `editTime`; the app writes it.

**Money / counts:** suffix `amount` / `price` / `count` / `num` / `quantity` when that is the meaning.

**YN flags:** `VARCHAR(1)` with `'Y'` / `'N'`.

---

## Common columns

Business tables include these. The business id column name depends on the table and usually forms a composite PK with `tenantId`.

| Column | Type | Notes |
|--------|------|--------|
| `<businessId>` | VARCHAR(32) | Not `id` |
| `tenantId` | VARCHAR(32) NOT NULL | Tenant isolation; usually part of the PK |
| `addTime` / `addWho` | DATETIME / VARCHAR(32) | Created |
| `editTime` / `editWho` | DATETIME / VARCHAR(32) | Last edit |
| `oprSeqFlag` | VARCHAR(32) | Operation sequence / optimistic lock |
| `currentVersion` | INT DEFAULT 1 | Increment on update |
| `activeFlag` | VARCHAR(1) DEFAULT `'Y'` | `'Y'` / `'N'` |
| `noteText` | VARCHAR(500) | Remark |
| `extProperty` | TEXT | JSON extras |
| `reserved1`… | text or typed | Count and types follow the `.sql` |

---

## Indexes

- Oracle names ≤ 30 characters; MySQL / SQLite ≤ 64
- Index pattern: `IDX_<table-abbrev>_<column-abbrev>` e.g. `IDX_GW_INST_BIND_HTTP`
- Always filter by `tenantId` in queries

---

## Examples

Full DDL lives under `scripts/db/{mysql,sqlite,oracle}/`.

- Users: `HUB_USER` (`scripts/db/sqlite/HUB_USER.sql`). PK `(userId, tenantId)`. There is no `HUB_USER_ACCOUNT`.
- Gateway instances: `HUB_GW_INSTANCE`. PK `(tenantId, gatewayInstanceId)`. Activity is `activeFlag`; health is `healthStatus`. There is no `instanceStatus` column.

Do not invent a second schema in docs. Change the scripts, then this page if the convention itself changed.

---

**[Index](./README.md) • [Previous: Container deployment](./04-container-deployment.md) • [Next: Debugging](./06-debugging.md)**
