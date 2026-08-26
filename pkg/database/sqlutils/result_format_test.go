package sqlutils

import (
	"reflect"
	"strings"
	"testing"
)

type countRow struct {
	Count int `db:"COUNT(*)"`
}

type namedCountRow struct {
	Total int `db:"total"`
}

type userRow struct {
	UserId   string `db:"userId"`
	UserName string `db:"userName"`
}

func TestFindFieldByColumnCountAliases(t *testing.T) {
	row := countRow{}
	v := reflect.ValueOf(&row).Elem()

	for _, col := range []string{"", "COUNT(*)", "cnt", "count", "(No column name)"} {
		field, ok := FindFieldByColumn(v, col)
		if !ok {
			t.Fatalf("column %q should match COUNT(*)", col)
		}
		if !field.CanSet() {
			t.Fatalf("column %q matched unsettable field", col)
		}
	}

	if _, ok := FindFieldByColumn(v, "userId"); ok {
		t.Fatal("userId should not match COUNT(*)")
	}
}

func TestPrepareScanTargetsEmptyCountColumn(t *testing.T) {
	row := countRow{}
	v := reflect.ValueOf(&row).Elem()
	targets, fields := PrepareScanTargetsWithFields(v, []string{""})
	if len(targets) != 1 || len(fields) != 1 {
		t.Fatalf("targets=%d fields=%d", len(targets), len(fields))
	}
	if !fields[0].IsValid() || !fields[0].CanSet() {
		t.Fatal("empty COUNT column should bind to Count")
	}
}

func TestPrepareScanTargetsDoesNotBindCountToUser(t *testing.T) {
	row := userRow{}
	v := reflect.ValueOf(&row).Elem()
	_, fields := PrepareScanTargetsWithFields(v, []string{""})
	if fields[0].IsValid() {
		t.Fatal("multi-field struct should not take unnamed column")
	}
}

func TestMapValuesCountAlias(t *testing.T) {
	row := namedCountRow{}
	mapper, err := NewFieldMapper([]string{""}, &row)
	if err != nil {
		t.Fatal(err)
	}
	if err := mapper.MapValues([]interface{}{int64(7)}); err != nil {
		t.Fatal(err)
	}
	if row.Total != 7 {
		t.Fatalf("Total=%d", row.Total)
	}
}

func TestBuildCountQueryAliasesCnt(t *testing.T) {
	got, err := BuildCountQuery("SELECT * FROM HUB_USER WHERE tenantId = ? ORDER BY addTime DESC")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.ToUpper(got), "SELECT COUNT(*) AS CNT ") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(strings.ToUpper(got), "ORDER BY") {
		t.Fatalf("ORDER BY should be stripped: %q", got)
	}
}
