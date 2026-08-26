package limiter

import (
	"os"
	"testing"
	"time"

	"gateway/pkg/cache"
	"gateway/pkg/cache/memory"
)

// TestMain 为限流单测挂上全局 default 缓存，与生产「只复用 Manager」一致。
func TestMain(m *testing.M) {
	if cache.GetDefaultCache() == nil {
		mem, err := memory.NewMemoryCache(&memory.MemoryConfig{
			Enabled:           true,
			EnableLazyCleanup: true,
			CleanupInterval:   time.Minute,
		})
		if err != nil {
			panic(err)
		}
		if err := cache.AddCache("default", mem); err != nil && cache.GetDefaultCache() == nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}
