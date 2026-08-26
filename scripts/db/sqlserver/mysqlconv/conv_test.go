package mysqlconv

import (
	"strings"
	"testing"
)

func TestConvertPreservesStringLiteralsTextAndJSON(t *testing.T) {
	in := "CREATE TABLE IF NOT EXISTS `T` (\n" +
		"  `contentType` VARCHAR(50) DEFAULT 'text',\n" +
		"  `logFormat` VARCHAR(50) NOT NULL DEFAULT 'JSON',\n" +
		"  `body` TEXT DEFAULT NULL\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='x';\n"
	out := Convert(in, "t.sql", nil)
	if strings.Contains(out, "DEFAULT N'NVARCHAR(MAX)'") || strings.Contains(out, "DEFAULT 'NVARCHAR(MAX)'") {
		t.Fatalf("string literal was rewritten:\n%s", out)
	}
	if !strings.Contains(out, "DEFAULT N'text'") {
		t.Fatalf("contentType default lost:\n%s", out)
	}
	if !strings.Contains(out, "DEFAULT N'JSON'") {
		t.Fatalf("logFormat default lost:\n%s", out)
	}
	if !strings.Contains(out, "body NVARCHAR(MAX)") {
		t.Fatalf("TEXT column not mapped:\n%s", out)
	}
	if !strings.Contains(out, "contentType NVARCHAR(50)") {
		t.Fatalf("VARCHAR not mapped to NVARCHAR:\n%s", out)
	}
}

func TestConvertInsertIgnoreValuesIsIdempotent(t *testing.T) {
	in := "INSERT IGNORE INTO `HUB_AUTH_RESOURCE` (\n" +
		"  `resourceId`, `tenantId`, `resourceName`\n" +
		") VALUES (\n" +
		"  'hub0009', 'default', '环境设置'\n" +
		");\n"
	out := Convert(in, "patch.sql", nil)
	if !strings.Contains(out, "IF NOT EXISTS") {
		t.Fatalf("missing IF NOT EXISTS:\n%s", out)
	}
	if !strings.Contains(out, "N'环境设置'") {
		t.Fatalf("missing unicode literal:\n%s", out)
	}
	if !strings.Contains(out, "ELSE") || !strings.Contains(out, "UPDATE HUB_AUTH_RESOURCE") {
		t.Fatalf("missing upsert update:\n%s", out)
	}
}

func TestConvertInsertIgnoreSelectNotExists(t *testing.T) {
	in := "INSERT IGNORE INTO `HUB_AUTH_ROLE_RESOURCE` (\n" +
		"  `roleResourceId`, `tenantId`, `roleId`, `resourceId`\n" +
		")\nSELECT\n  CONCAT('ROLE_RES_VIEWER_', REPLACE(`resourceId`, ':', '_')),\n" +
		"  `tenantId`,\n  'ROLE_VIEWER',\n  `resourceId`\n" +
		"FROM `HUB_AUTH_RESOURCE`\nWHERE `tenantId` = 'default';\n"
	out := Convert(in, "patch.sql", nil)
	if !strings.Contains(out, "AND NOT EXISTS") {
		t.Fatalf("missing NOT EXISTS:\n%s", out)
	}
	if !strings.Contains(out, "N'ROLE_VIEWER'") {
		t.Fatalf("role literal not unicode:\n%s", out)
	}
}

func TestServiceDefinitionIdStaysVarchar(t *testing.T) {
	in := "CREATE TABLE `T` (\n  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL,\n  `routeName` VARCHAR(100) NOT NULL\n);\n"
	out := Convert(in, "t.sql", nil)
	if !strings.Contains(out, "serviceDefinitionId VARCHAR(32)") {
		t.Fatalf("serviceDefinitionId should stay VARCHAR:\n%s", out)
	}
	if !strings.Contains(out, "routeName NVARCHAR(100)") {
		t.Fatalf("routeName should be NVARCHAR:\n%s", out)
	}
	if !strings.Contains(out, "ALTER TABLE T ALTER COLUMN routeName NVARCHAR(100) NOT NULL") {
		t.Fatalf("missing unicode alter for existing DBs:\n%s", out)
	}
}

func TestAddColumnGuarded(t *testing.T) {
	in := "ALTER TABLE `HUB_USER` ADD COLUMN `mustChangePwd` VARCHAR(1) NOT NULL DEFAULT 'N';\n"
	out := Convert(in, "HUB_USER.sql", nil)
	if !strings.Contains(out, "IF COL_LENGTH(N'dbo.HUB_USER', N'mustChangePwd') IS NULL") {
		t.Fatalf("missing COL_LENGTH guard:\n%s", out)
	}
}

func TestUserSeedNoPasswordReset(t *testing.T) {
	in := "INSERT INTO HUB_USER (\n    userId,\n    tenantId,\n    realName\n) VALUES (\n    'admin', -- userId\n    'default', -- tenantId\n    '系统管理员'\n);\n"
	out := Convert(in, "HUB_USER.sql", nil)
	if !strings.Contains(out, "IF NOT EXISTS") {
		t.Fatalf("user seed not guarded:\n%s", out)
	}
	if strings.Contains(out, "password =") {
		t.Fatalf("user seed must not UPDATE password:\n%s", out)
	}
	if !strings.Contains(out, "ELSE") || !strings.Contains(out, "realName = N'系统管理员'") {
		t.Fatalf("user seed should repair realName:\n%s", out)
	}
	if !strings.Contains(out, "N'系统管理员'") {
		t.Fatalf("missing unicode name:\n%s", out)
	}
}

func TestInitSkipsMissingFiles(t *testing.T) {
	in := ":r HUB_USER.sql;\n:r HUB_REGISTRY_SERVICE.sql;\n"
	out := Convert(in, "init.sql", map[string]struct{}{"HUB_USER.sql": {}})
	if !strings.Contains(out, ":r HUB_USER.sql") {
		t.Fatalf("kept file dropped:\n%s", out)
	}
	if strings.Contains(out, ":r HUB_REGISTRY_SERVICE.sql") {
		t.Fatalf("missing file not skipped:\n%s", out)
	}
}
