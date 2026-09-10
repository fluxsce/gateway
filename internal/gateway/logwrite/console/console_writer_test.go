package console

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"gateway/internal/gateway/logwrite/types"
)

func testConsoleConfig(queue int) *types.LogConfig {
	cfg := &types.LogConfig{
		LogFormat:          string(types.LogFormatJSON),
		EnableAsyncLogging: "Y",
		AsyncQueueSize:     queue,
		OutputTargets:      string(types.LogOutputConsole),
	}
	cfg.SetDefaults()
	cfg.EnableAsyncLogging = "Y"
	cfg.AsyncQueueSize = queue
	return cfg
}

func TestConsoleWriterDropsWhenStdoutBlocked(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	cw, err := NewConsoleWriter(testConsoleConfig(100))
	if err != nil {
		t.Fatal(err)
	}
	cw.output = w
	t.Cleanup(func() { _ = cw.Close() })

	log := types.NewAccessLog("t1", "gw1", "127.0.0.1")
	log.RequestPath = "/x"
	ctx := context.Background()
	for i := 0; i < 500; i++ {
		if err := cw.Write(ctx, log); err != nil {
			t.Fatalf("Write must not fail when dropping: %v", err)
		}
	}
	if cw.dropped.Load() == 0 {
		t.Fatal("blocked stdout must drop after the queue fills")
	}
}

func TestConsoleWriterWritesWhenDrained(t *testing.T) {
	var n atomic.Int64
	cw, err := NewConsoleWriter(testConsoleConfig(100))
	if err != nil {
		t.Fatal(err)
	}
	cw.output = writeFunc(func(p []byte) (int, error) {
		n.Add(1)
		return len(p), nil
	})
	t.Cleanup(func() { _ = cw.Close() })

	log := types.NewAccessLog("t1", "gw1", "127.0.0.1")
	if err := cw.Write(context.Background(), log); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for n.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n.Load() == 0 {
		t.Fatal("expected a console line to be written")
	}
}

type writeFunc func([]byte) (int, error)

func (f writeFunc) Write(p []byte) (int, error) { return f(p) }
