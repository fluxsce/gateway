package dialect

import (
	"strings"
	"testing"

	"gateway/pkg/database/dbtypes"
)

func TestGenerateOracleOmitsCharsetWhenEmpty(t *testing.T) {
	got, err := generateOracle(&dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Username: "u",
			Password: "p",
			Database: "XE",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"CHARSET=", "NLS_CHARACTERSET=", "NLS_LANG="} {
		if strings.Contains(got, needle) {
			t.Fatalf("未配置时不应写 %s，实际 %s", needle, got)
		}
	}
}

func TestGenerateOracleWritesCharsetWhenSet(t *testing.T) {
	got, err := generateOracle(&dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:            "localhost",
			Username:        "u",
			Password:        "p",
			Database:        "XE",
			Charset:         "UTF8",
			NLSLang:         "CHINESE_CHINA.AL32UTF8",
			NLSCharacterset: "AL32UTF8",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"CHARSET=UTF8", "NLS_CHARACTERSET=AL32UTF8", "NLS_LANG=CHINESE_CHINA.AL32UTF8"} {
		if !strings.Contains(got, needle) {
			t.Fatalf("显式配置应写入 %s，实际 %s", needle, got)
		}
	}
}
