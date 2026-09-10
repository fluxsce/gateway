package logger

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestExtraConsoleSink(t *testing.T) {
	if extraConsoleSink("console") != "stdout" {
		t.Fatal("output=console should write stdout for collection")
	}
	if extraConsoleSink("stdout") != "" || extraConsoleSink("stderr") != "" {
		t.Fatal("output=stdout/stderr should stay original and not add a console sink")
	}
}

func TestConfiguredSinksKeepsLegacyFiles(t *testing.T) {
	got := configuredSinks(&LoggerConfig{
		DefaultOutput: "gateway.log",
		ErrorOutput:   "error.log",
		WarnOutput:    "warn.log",
		InfoOutput:    "info.log",
		DebugOutput:   "debug.log",
	})
	want := []string{"gateway.log", "error.log", "warn.log", "info.log", "debug.log"}
	if len(got) != len(want) {
		t.Fatalf("legacy files = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("legacy files = %#v", got)
		}
	}
}

func TestConfiguredSinksConsoleIsExtra(t *testing.T) {
	got := configuredSinks(&LoggerConfig{
		Output:        "console",
		DefaultOutput: "gateway.log",
		ErrorOutput:   "error.log",
	})
	if len(got) != 3 || got[0] != "gateway.log" || got[1] != "error.log" || got[2] != "stdout" {
		t.Fatalf("console should be extra, not replace files: %#v", got)
	}
}

func TestConfiguredSinksStdoutOnOutputIsIgnored(t *testing.T) {
	got := configuredSinks(&LoggerConfig{Output: "stdout", DefaultOutput: "gateway.log"})
	if len(got) != 1 || got[0] != "gateway.log" {
		t.Fatalf("output=stdout must not add console: %#v", got)
	}
}

func TestConfiguredSinksNoDuplicateStdout(t *testing.T) {
	got := configuredSinks(&LoggerConfig{Output: "console", DefaultOutput: "stdout"})
	if len(got) != 1 || got[0] != "stdout" {
		t.Fatalf("stdout should not be added twice: %#v", got)
	}
}

func TestRenderJSONFormatBindsVars(t *testing.T) {
	ent := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC),
		Message: "hello",
	}
	line, err := renderFormat(
		`{"time":"${time}","level":"${level}","msg":"${msg}"}`,
		"json",
		"rfc3339",
		"",
		ent,
		[]zapcore.Field{{Key: "requestId", Type: zapcore.StringType, String: "r1"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("line=%s err=%v", line, err)
	}
	if obj["level"] != "info" || obj["msg"] != "hello" || obj["requestId"] != "r1" {
		t.Fatalf("bound json = %#v", obj)
	}
	if !strings.HasPrefix(obj["time"].(string), "2026-09-10") {
		t.Fatalf("time = %v", obj["time"])
	}
}

func TestRenderTextFormatBindsVars(t *testing.T) {
	ent := zapcore.Entry{Level: zapcore.WarnLevel, Message: "slow", Time: time.Unix(0, 0).UTC()}
	line, err := renderFormat(`${level} ${msg}`, "console", "epoch", "capital", ent, nil)
	if err != nil {
		t.Fatal(err)
	}
	if line != "WARN slow" {
		t.Fatalf("text = %q", line)
	}
}

func TestInitWithFormat(t *testing.T) {
	if err := Init(&LoggerConfig{
		Level:         "info",
		Encoding:      "json",
		Format:        `{"time":"${time}","level":"${level}","msg":"${msg}"}`,
		Output:        "stdout",
		DefaultOutput: "stdout",
		TimeFormat:    "iso8601",
		ShowCaller:    false,
	}); err != nil {
		t.Fatalf("Init format: %v", err)
	}
}
