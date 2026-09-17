package asyncq

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestOfferSucceedsWhenCapacity(t *testing.T) {
	ch := make(chan int, 1)
	if !Offer(ch, 7) {
		t.Fatal("有空位时应入队成功")
	}
	if got := <-ch; got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestOfferWaitZeroDropsWhenFull(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	if OfferWait(ch, 2, 0) {
		t.Fatal("满且 wait=0 应立刻失败")
	}
}

func TestOfferWaitsThenSucceeds(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	go func() {
		time.Sleep(5 * time.Millisecond)
		<-ch
	}()
	if !OfferWait(ch, 2, 50*time.Millisecond) {
		t.Fatal("短等后应能入队")
	}
}

func TestDrain(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	sum := 0
	n := Drain(ch, func(v int) { sum += v })
	if n != 2 || sum != 3 {
		t.Fatalf("n=%d sum=%d, want 2 and 3", n, sum)
	}
}

func TestRestorePrependAndAppendCapped(t *testing.T) {
	var mu sync.Mutex
	buf := []int{3, 4}
	dropped := RestorePrepend(&mu, &buf, []int{1, 2}, 3, 3)
	if dropped != 1 || len(buf) != 3 || buf[0] != 1 || buf[1] != 2 || buf[2] != 3 {
		t.Fatalf("RestorePrepend buf=%v dropped=%d, want [1 2 3] dropped=1", buf, dropped)
	}
	if AppendCapped(&mu, &buf, 9, 3) != 1 {
		t.Fatal("已满应拒绝新条目")
	}
	buf = buf[:2]
	if AppendCapped(&mu, &buf, 8, 3) != 0 || buf[2] != 8 {
		t.Fatalf("未满应追加, buf=%v", buf)
	}
}

func TestRetryableWriteError(t *testing.T) {
	if !RetryableWriteError(context.DeadlineExceeded) {
		t.Fatal("deadline 应可重试")
	}
	if !RetryableWriteError(fmt.Errorf("mongo insert error (2103): server selection timeout")) {
		t.Fatal("server selection 应可重试")
	}
	if RetryableWriteError(fmt.Errorf("E11000 duplicate key")) {
		t.Fatal("重复键不可回灌")
	}
	if RetryableWriteError(nil) {
		t.Fatal("nil 不可重试")
	}
}

func TestTakeAndAppendTakeIfFull(t *testing.T) {
	var mu sync.Mutex
	buf := make([]int, 0, 2)
	if got := AppendTakeIfFull(&mu, &buf, 1, 2, 2); got != nil {
		t.Fatal("未满不应拿走")
	}
	got := AppendTakeIfFull(&mu, &buf, 2, 2, 2)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("满批 = %v, want [1 2]", got)
	}
	if Take(&mu, &buf, 2) != nil {
		t.Fatal("拿走后应为空")
	}
	buf = append(buf, 9)
	if got := Take(&mu, &buf, 2); len(got) != 1 || got[0] != 9 {
		t.Fatalf("Take = %v, want [9]", got)
	}
}

func TestEnsureWriteCtxReusesDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx, done := EnsureWriteCtx(parent)
	defer done()
	if ctx != parent {
		t.Fatal("已有 deadline 时应沿用原 context")
	}
}

func TestEnsureWriteCtxAddsTimeout(t *testing.T) {
	ctx, cancel := EnsureWriteCtx(context.Background())
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("无 deadline 时应套上写超时")
	}
}

func TestWriteContextCancel(t *testing.T) {
	ctx, cancel := WriteContext(time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("WriteContext 应在超时后取消")
	}
}

func TestEnsureWriteCtxTimeoutUsesConfig(t *testing.T) {
	ctx, cancel := EnsureWriteCtxTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	dl, ok := ctx.Deadline()
	if !ok {
		t.Fatal("无 deadline 时应套上配置超时")
	}
	remain := time.Until(dl)
	if remain > 20*time.Millisecond || remain < 0 {
		t.Fatalf("deadline 应接近 10ms, remain=%s", remain)
	}
}
