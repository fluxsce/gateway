package bootstrap

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"gateway/pkg/logger"
)

// listenerDispatcher 持有唯一的操作系统监听器，并把新连接投递到当前Server代际。
type listenerDispatcher struct {
	base net.Listener

	mu      sync.RWMutex
	current *virtualListener

	runOnce   sync.Once
	closeOnce sync.Once
	done      chan struct{}
	runDone   chan struct{}

	maxConnections        atomic.Int64
	activeConnections     atomic.Int64
	rejectedConnections   atomic.Uint64
	lastRejectLogUnixNano atomic.Int64
}

// newListenerDispatcher 创建连接分发器。
func newListenerDispatcher(base net.Listener) *listenerDispatcher {
	return &listenerDispatcher{
		base:    base,
		done:    make(chan struct{}),
		runDone: make(chan struct{}),
	}
}

// start 幂等启动底层连接接收循环。
func (d *listenerDispatcher) start() {
	d.runOnce.Do(func() {
		go d.run()
	})
}

// run 持续接收底层连接，并确保切换前已接收的连接仍归属旧代际。
func (d *listenerDispatcher) run() {
	defer close(d.runDone)
	var retryDelay time.Duration
	for {
		conn, err := d.base.Accept()
		if err != nil {
			var stop bool
			retryDelay, stop = d.waitAcceptRetry(err, retryDelay)
			if stop {
				return
			}
			continue
		}
		retryDelay = 0
		conn = d.trackAcceptedConnection(conn)
		if conn == nil {
			continue
		}

		d.mu.RLock()
		target := d.current
		if target != nil {
			target.deliveries.Add(1)
		}
		d.mu.RUnlock()
		// 投递时不持有分发器锁，避免队列阻塞导致Reload无法取得写锁。
		delivered := false
		if target != nil {
			delivered = target.deliver(conn, d.done)
			target.deliveries.Done()
		}
		if !delivered {
			_ = conn.Close()
		}
	}
}

// trackAcceptedConnection 校验连接容量，并用可自动释放计数的连接包装原始连接。
func (d *listenerDispatcher) trackAcceptedConnection(conn net.Conn) net.Conn {
	if !d.tryAcquireConnection() {
		_ = conn.Close()
		d.logConnectionRejection()
		return nil
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}
	return &trackedConn{
		Conn:    conn,
		release: d.releaseConnection,
	}
}

func (d *listenerDispatcher) waitAcceptRetry(err error, currentDelay time.Duration) (time.Duration, bool) {
	select {
	case <-d.done:
		return currentDelay, true
	default:
	}

	retryDelay := nextAcceptRetryDelay(currentDelay)
	logger.Warn("连接分发器接收连接失败，将延迟重试",
		"error", err, "retryDelay", retryDelay)
	timer := time.NewTimer(retryDelay)
	defer timer.Stop()
	select {
	case <-d.done:
		return retryDelay, true
	case <-timer.C:
		return retryDelay, false
	}
}

// newGenerationListener 创建共享同一监听地址的虚拟listener。
func (d *listenerDispatcher) newGenerationListener() *virtualListener {
	return &virtualListener{
		addr:    d.base.Addr(),
		conns:   make(chan net.Conn),
		done:    make(chan struct{}),
		readyCh: make(chan struct{}),
	}
}

// switchTo 原子切换新连接的目标代际，并返回旧虚拟listener。
func (d *listenerDispatcher) switchTo(next *virtualListener) *virtualListener {
	d.mu.Lock()
	old := d.current
	d.current = next
	d.mu.Unlock()
	// 写锁保证旧目标不会再新增投递，等待已Accept的连接完成归属后再排空旧Server。
	if old != nil {
		old.deliveries.Wait()
	}
	return old
}

// Close 停止接收新连接；只有该方法会关闭底层操作系统listener。
func (d *listenerDispatcher) Close() error {
	var closeErr error
	d.closeOnce.Do(func() {
		close(d.done)
		closeErr = d.base.Close()
	})
	// 即使接收循环尚未启动，也启动一次以观察关闭状态并关闭runDone。
	d.start()
	<-d.runDone
	return closeErr
}

func nextAcceptRetryDelay(current time.Duration) time.Duration {
	const (
		initialDelay = 5 * time.Millisecond
		maxDelay     = time.Second
	)
	if current <= 0 {
		return initialDelay
	}
	if current >= maxDelay/2 {
		return maxDelay
	}
	return current * 2
}

func (d *listenerDispatcher) setMaxConnections(limit int) {
	d.maxConnections.Store(int64(limit))
}

func (d *listenerDispatcher) tryAcquireConnection() bool {
	for {
		current := d.activeConnections.Load()
		limit := d.maxConnections.Load()
		if limit > 0 && current >= limit {
			d.rejectedConnections.Add(1)
			return false
		}
		if d.activeConnections.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func (d *listenerDispatcher) releaseConnection() {
	d.activeConnections.Add(-1)
}

func (d *listenerDispatcher) logConnectionRejection() {
	now := time.Now().UnixNano()
	last := d.lastRejectLogUnixNano.Load()
	if now-last < int64(time.Second) ||
		!d.lastRejectLogUnixNano.CompareAndSwap(last, now) {
		return
	}
	logger.Warn("网关并发连接数已达到上限，拒绝新连接",
		"activeConnections", d.activeConnections.Load(),
		"maxConnections", d.maxConnections.Load(),
		"rejectedConnections", d.rejectedConnections.Load())
}

// virtualListener 为单个http.Server提供隔离的Accept和Close语义。
type virtualListener struct {
	addr net.Addr

	conns   chan net.Conn
	done    chan struct{}
	once    sync.Once
	ready   sync.Once
	readyCh chan struct{}

	deliveries sync.WaitGroup
}

func (l *virtualListener) deliver(conn net.Conn, dispatcherDone <-chan struct{}) bool {
	select {
	case <-dispatcherDone:
		return false
	case <-l.done:
		return false
	case l.conns <- conn:
		return true
	}
}

// Accept 返回分配给当前代际的新连接。
func (l *virtualListener) Accept() (net.Conn, error) {
	l.ready.Do(func() {
		close(l.readyCh)
	})
	select {
	case <-l.done:
		return nil, net.ErrClosed
	case conn := <-l.conns:
		if conn == nil {
			return nil, net.ErrClosed
		}
		return conn, nil
	}
}

// Close 停止该代际接收连接，但不会关闭共享的底层监听端口。
func (l *virtualListener) Close() error {
	l.once.Do(func() {
		close(l.done)
	})
	return nil
}

// Addr 返回共享底层listener的监听地址。
func (l *virtualListener) Addr() net.Addr {
	return l.addr
}

// trackedConn 在连接实际关闭时释放全局并发连接名额。
type trackedConn struct {
	net.Conn
	closeOnce sync.Once
	closeErr  error
	release   func()
}

// Close 关闭底层连接并确保连接计数只释放一次。
func (c *trackedConn) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.Conn.Close()
		if c.release != nil {
			c.release()
		}
	})
	return c.closeErr
}

var _ net.Listener = (*virtualListener)(nil)
var _ net.Conn = (*trackedConn)(nil)
