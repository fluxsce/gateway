package controllers

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"gateway/web/views/hub0004/models"
)

func TestWriteAuditLogCSV(t *testing.T) {
	rows := []*models.AuthAuditLog{
		{
			UserName:      `张"三`,
			UserId:        "u1",
			Action:        "CREATE",
			ModuleCode:    "hub0002",
			TargetType:    "USER",
			TargetName:    "a,b",
			TargetId:      "id-1",
			ResourceCode:  "hub0002:add",
			Result:        "Y",
			ClientIP:      "127.0.0.1",
			RequestMethod: "POST",
			RequestPath:   "/gateway/hub0002/addUser",
			Detail:        "line1\nline2",
			AddTime:       time.Date(2026, 8, 20, 21, 0, 0, 0, time.Local),
		},
	}
	var buf bytes.Buffer
	if err := writeAuditLogCSV(&buf, rows); err != nil {
		t.Fatalf("writeAuditLogCSV: %v", err)
	}
	payload := buf.Bytes()
	if !bytes.HasPrefix(payload, utf8BOM) {
		t.Fatal("missing UTF-8 BOM")
	}
	r := csv.NewReader(bytes.NewReader(payload[len(utf8BOM):]))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("want 2 records, got %d", len(records))
	}
	if records[1][1] != `张"三` || records[1][6] != "a,b" {
		t.Fatalf("escaped fields mismatch: %#v", records[1])
	}
	if !strings.Contains(records[1][13], "line1") {
		t.Fatalf("detail lost: %q", records[1][13])
	}
}

func TestAuditExportFilename(t *testing.T) {
	got := auditExportFilename(time.Date(2026, 8, 20, 21, 30, 0, 0, time.Local))
	if got != "audit-log-20260820213000.csv" {
		t.Fatalf("got %s", got)
	}
}
