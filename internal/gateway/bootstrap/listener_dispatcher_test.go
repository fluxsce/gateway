package bootstrap

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestListenerDispatcherCloseBeforeStart(t *testing.T) {
	base, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newListenerDispatcher(base)

	done := make(chan error, 1)
	go func() {
		done <- dispatcher.Close()
	}()

	select {
	case closeErr := <-done:
		if closeErr != nil && closeErr != net.ErrClosed {
			t.Fatal(closeErr)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() blocked before dispatcher start")
	}
}

func TestListenerDispatcherBacksOffAcceptErrors(t *testing.T) {
	base := newFailingListener()
	dispatcher := newListenerDispatcher(base)
	dispatcher.start()

	select {
	case <-base.firstAccept:
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not call Accept()")
	}
	time.Sleep(50 * time.Millisecond)

	if calls := base.acceptCalls.Load(); calls < 2 || calls > 6 {
		t.Fatalf("Accept() calls during retry window = %d, want 2..6", calls)
	}
	if err := dispatcher.Close(); err != nil && err != net.ErrClosed {
		t.Fatal(err)
	}
}

// failingListener 模拟持续Accept失败的底层监听器，用于验证错误退避和关闭行为。
type failingListener struct {
	acceptCalls atomic.Int64
	firstOnce   sync.Once
	firstAccept chan struct{}
	closeOnce   sync.Once
	closed      chan struct{}
}

func newFailingListener() *failingListener {
	return &failingListener{
		firstAccept: make(chan struct{}),
		closed:      make(chan struct{}),
	}
}

func (l *failingListener) Accept() (net.Conn, error) {
	l.firstOnce.Do(func() {
		close(l.firstAccept)
	})
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
		l.acceptCalls.Add(1)
		return nil, errors.New("temporary accept failure")
	}
}

func (l *failingListener) Close() error {
	l.closeOnce.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *failingListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

var _ net.Listener = (*failingListener)(nil)
