package logger

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

var placeholderRe = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)

// formatCore 按 format 模板写出一行，结构化字段通过 logger.With / 调用参数并入。
type formatCore struct {
	ws       zapcore.WriteSyncer
	enab     zapcore.LevelEnabler
	format   string
	encoding string
	timeFmt  string
	levelEnc string
	stored   []zapcore.Field
}

// newFormatCore 创建模板输出核心。format 为空时不应调用。
func newFormatCore(cfg *LoggerConfig, ws zapcore.WriteSyncer, enab zapcore.LevelEnabler) zapcore.Core {
	return &formatCore{
		ws:       ws,
		enab:     enab,
		format:   cfg.Format,
		encoding: cfg.Encoding,
		timeFmt:  cfg.TimeFormat,
		levelEnc: cfg.LevelEncoder,
	}
}

func (c *formatCore) Enabled(level zapcore.Level) bool {
	return c.enab.Enabled(level)
}

func (c *formatCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.stored = append(append([]zapcore.Field{}, c.stored...), fields...)
	return &clone
}

func (c *formatCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *formatCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	all := make([]zapcore.Field, 0, len(c.stored)+len(fields))
	all = append(all, c.stored...)
	all = append(all, fields...)
	line, err := renderFormat(c.format, c.encoding, c.timeFmt, c.levelEnc, ent, all)
	if err != nil {
		return err
	}
	if line == "" || line[len(line)-1] != '\n' {
		line += "\n"
	}
	_, err = c.ws.Write([]byte(line))
	if err != nil {
		return err
	}
	if ent.Level >= zapcore.ErrorLevel {
		return c.Sync()
	}
	return nil
}

func (c *formatCore) Sync() error {
	return c.ws.Sync()
}

// renderFormat 用 ${name} 绑定条目字段。JSON 对象模板会并入其余结构化字段。
func renderFormat(format, encoding, timeFmt, levelEnc string, ent zapcore.Entry, fields []zapcore.Field) (string, error) {
	vars := entryVars(ent, timeFmt, levelEnc)
	trim := strings.TrimSpace(format)
	if strings.EqualFold(encoding, "json") || strings.HasPrefix(trim, "{") {
		return renderJSONFormat(trim, vars, fields)
	}
	return renderTextFormat(format, vars, fields), nil
}

func entryVars(ent zapcore.Entry, timeFmt, levelEnc string) map[string]string {
	caller := ""
	if ent.Caller.Defined {
		caller = ent.Caller.TrimmedPath()
	}
	return map[string]string{
		"time":   formatEntryTime(ent.Time, timeFmt),
		"level":  formatEntryLevel(ent.Level, levelEnc),
		"msg":    ent.Message,
		"caller": caller,
		"stack":  ent.Stack,
		"logger": ent.LoggerName,
	}
}

func formatEntryTime(t time.Time, format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "iso8601":
		return t.Format("2006-01-02T15:04:05.000Z07:00")
	case "rfc3339":
		return t.Format(time.RFC3339)
	case "rfc3339nano":
		return t.Format(time.RFC3339Nano)
	case "epoch":
		return strconv.FormatInt(t.Unix(), 10)
	case "epoch_millis":
		return strconv.FormatInt(t.UnixMilli(), 10)
	default:
		return t.Format(format)
	}
}

func formatEntryLevel(level zapcore.Level, encoderName string) string {
	name := level.String()
	switch strings.ToLower(strings.TrimSpace(encoderName)) {
	case "capital":
		return strings.ToUpper(name)
	default:
		return name
	}
}

func renderJSONFormat(format string, vars map[string]string, fields []zapcore.Field) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(format), &raw); err != nil {
		return renderTextFormat(format, vars, fields), nil
	}
	bound := bindJSON(raw, vars)
	if obj, ok := bound.(map[string]interface{}); ok {
		mergeFields(obj, fields)
	}
	b, err := json.Marshal(bound)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func bindJSON(v interface{}, vars map[string]string) interface{} {
	switch t := v.(type) {
	case string:
		return bindString(t, vars)
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, child := range t {
			out[k] = bindJSON(child, vars)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, child := range t {
			out[i] = bindJSON(child, vars)
		}
		return out
	default:
		return v
	}
}

func bindString(s string, vars map[string]string) string {
	if m := placeholderRe.FindStringSubmatch(s); len(m) == 2 && m[0] == s {
		if v, ok := lookupVar(m[1], vars); ok {
			return v
		}
		return ""
	}
	return expandPlaceholders(s, vars)
}

func expandPlaceholders(s string, vars map[string]string) string {
	return placeholderRe.ReplaceAllStringFunc(s, func(m string) string {
		name := m[2 : len(m)-1]
		if v, ok := lookupVar(name, vars); ok {
			return v
		}
		return ""
	})
}

func lookupVar(name string, vars map[string]string) (string, bool) {
	if v, ok := vars[strings.ToLower(name)]; ok {
		return v, true
	}
	return "", false
}

func renderTextFormat(format string, vars map[string]string, fields []zapcore.Field) string {
	line := expandPlaceholders(format, vars)
	if strings.Contains(format, "${fields}") {
		line = strings.ReplaceAll(line, "${fields}", fieldsText(fields))
	} else if extra := fieldsText(fields); extra != "" {
		line = strings.TrimRight(line, " ") + extra
	}
	return line
}

func mergeFields(obj map[string]interface{}, fields []zapcore.Field) {
	if len(fields) == 0 {
		return
	}
	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}
	for k, v := range enc.Fields {
		if _, exists := obj[k]; !exists {
			obj[k] = v
		}
	}
}

func fieldsText(fields []zapcore.Field) string {
	if len(fields) == 0 {
		return ""
	}
	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}
	if len(enc.Fields) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range enc.Fields {
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(stringifyField(v))
	}
	return b.String()
}

func stringifyField(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
