# FLUX Gateway - Database Design Specifications

This document provides database design specifications and best practices for FLUX Gateway.

---

## 1. Naming Conventions

### Table Naming

- Use **lowercase** with **underscores** to separate words
- Use **singular** nouns (e.g., `user` not `users`)
- Prefix with module name (e.g., `gateway_route`, `tunnel_server`)
- Avoid reserved keywords

**Examples:**
```sql
✅ Good: gateway_route, tunnel_server, user_role
❌ Bad: GatewayRoute, TunnelServers, user-role, order (reserved keyword)
```

### Column Naming

- Use **lowercase** with **underscores**
- Use **descriptive** names
- Boolean fields prefix with `is_` or `has_`
- Time fields suffix with `_time` or `_at`

**Examples:**
```sql
✅ Good: user_name, is_active, created_time, updated_at
❌ Bad: UserName, active, create, update
```

### Index Naming

- **Primary Key**: `pk_<table_name>`
- **Unique Index**: `uk_<table_name>_<column_name>`
- **Normal Index**: `idx_<table_name>_<column_name>`
- **Foreign Key**: `fk_<table_name>_<ref_table>`

**Examples:**
```sql
✅ Good: pk_user, uk_user_email, idx_user_status, fk_order_user
❌ Bad: user_pk, email_unique, status_idx
```

---

## 2. Required Common Fields

All tables **MUST** include the following fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `id` | BIGINT / VARCHAR(36) | AUTO_INCREMENT / UUID | Primary key |
| `create_time` | DATETIME / TIMESTAMP | CURRENT_TIMESTAMP | Creation time |
| `update_time` | DATETIME / TIMESTAMP | CURRENT_TIMESTAMP ON UPDATE | Last update time |
| `is_deleted` | TINYINT(1) | 0 | Soft delete flag (0: not deleted, 1: deleted) |

**Example:**
```sql
CREATE TABLE gateway_route (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Primary key ID',
    route_name VARCHAR(100) NOT NULL COMMENT 'Route name',
    route_path VARCHAR(255) NOT NULL COMMENT 'Route path',
    
    -- Required common fields
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Update time',
    is_deleted TINYINT(1) DEFAULT 0 COMMENT 'Soft delete flag'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gateway route table';
```

---

## 3. Field Naming Semantic Requirements

### Status Fields

Use **clear** status values:

```sql
-- ✅ Good: Use meaningful enum values
status ENUM('active', 'inactive', 'pending', 'disabled') DEFAULT 'active'

-- ❌ Bad: Use unclear numeric codes
status TINYINT DEFAULT 1  -- What does 1 mean?
```

### Boolean Fields

Use `is_` or `has_` prefix:

```sql
-- ✅ Good
is_active TINYINT(1) DEFAULT 1
has_permission TINYINT(1) DEFAULT 0
is_deleted TINYINT(1) DEFAULT 0

-- ❌ Bad
active INT DEFAULT 1
permission BOOLEAN
deleted BIT
```

### Time Fields

Use `_time` or `_at` suffix:

```sql
-- ✅ Good
create_time DATETIME
updated_at TIMESTAMP
deleted_time DATETIME
expired_at DATETIME

-- ❌ Bad
create DATETIME
update TIMESTAMP
delete_date DATE
```

---

## 4. Other Specifications

### Data Types

| Use Case | Recommended Type | Notes |
|----------|------------------|-------|
| **Primary Key** | BIGINT AUTO_INCREMENT | Or VARCHAR(36) for UUID |
| **String** | VARCHAR(N) | Specify appropriate length |
| **Long Text** | TEXT / LONGTEXT | For articles, descriptions |
| **Boolean** | TINYINT(1) | 0 or 1 |
| **Date/Time** | DATETIME | Or TIMESTAMP for auto-update |
| **Decimal** | DECIMAL(M,D) | For currency, precise calculations |
| **JSON** | JSON | For flexible data structures |

### Indexing

- **Primary Key**: Always on `id`
- **Unique Index**: On unique fields (e.g., email, username)
- **Normal Index**: On frequently queried fields (e.g., status, type)
- **Composite Index**: For multi-column queries (order matters!)

**Example:**
```sql
-- Primary key
PRIMARY KEY (id),

-- Unique index
UNIQUE KEY uk_user_email (email),

-- Normal index
KEY idx_user_status (status),

-- Composite index (query by tenant_id and status)
KEY idx_tenant_status (tenant_id, status)
```

### Foreign Keys

- Use foreign keys for referential integrity
- Set appropriate `ON DELETE` and `ON UPDATE` actions

```sql
CONSTRAINT fk_order_user 
    FOREIGN KEY (user_id) 
    REFERENCES user(id) 
    ON DELETE CASCADE 
    ON UPDATE CASCADE
```

### Comments

- **Table Comment**: Describe table purpose
- **Column Comment**: Describe column meaning
- Use **Chinese** or **English** consistently

```sql
CREATE TABLE gateway_route (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Primary key ID',
    route_name VARCHAR(100) NOT NULL COMMENT 'Route name',
    route_path VARCHAR(255) NOT NULL COMMENT 'Route path',
    status ENUM('active', 'inactive') DEFAULT 'active' COMMENT 'Route status'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gateway route table';
```

---

## 5. Example

### Complete Table Example

```sql
CREATE TABLE gateway_route (
    -- Primary key
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'Primary key ID',
    
    -- Business fields
    route_name VARCHAR(100) NOT NULL COMMENT 'Route name',
    route_path VARCHAR(255) NOT NULL COMMENT 'Route path',
    target_url VARCHAR(500) NOT NULL COMMENT 'Target URL',
    method ENUM('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS') DEFAULT 'GET' COMMENT 'HTTP method',
    status ENUM('active', 'inactive', 'testing') DEFAULT 'active' COMMENT 'Route status',
    priority INT DEFAULT 0 COMMENT 'Priority (higher value = higher priority)',
    
    -- Optional fields
    description TEXT COMMENT 'Route description',
    config JSON COMMENT 'Route configuration (JSON)',
    
    -- Required common fields
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Update time',
    is_deleted TINYINT(1) DEFAULT 0 COMMENT 'Soft delete flag (0: not deleted, 1: deleted)',
    
    -- Indexes
    UNIQUE KEY uk_route_path (route_path),
    KEY idx_route_status (status),
    KEY idx_route_priority (priority),
    KEY idx_route_deleted (is_deleted)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gateway route table';
```

---

## 📖 Next Steps

After understanding database specifications, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Development Guide](./02-quick-start.md) - Development environment setup and configuration
- [Debugging Guide](./06-debugging.md) - Debugging techniques and troubleshooting

---

**[Back to Directory](./README.md) • [Previous: Containerized Deployment](./04-container-deployment.md) • [Next: Debugging Guide](./06-debugging.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

