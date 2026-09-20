package logwrite

import (
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/logwrite/asyncq"
	"gateway/pkg/logger"
)

// backendTraceJob 后端追踪提交任务。
// 入队前已抽好 ID、响应大小，以及过滤后真实转发的方法/URL/头/体；worker 不再读 ctx.Request / Writer。
// 头和体是调用方在代理 defer 里拷出的快照，不是活的 *http.Request。
type backendTraceJob struct {
	instanceID       string
	tenantID         string
	traceID          string
	serviceID        string
	serviceName      string
	requestMethod    string
	requestURL       string
	fallbackPath     string
	fallbackQuery    string
	requestSize      int
	responseSize     int
	requestStartTime time.Time
	responseTime     time.Time
	statusCode       int
	responseHeaders  map[string][]string
	responseBody     []byte
	forwardHeaders   map[string][]string
	forwardBody      []byte
	errText          string
	hasErr           bool
	retryCount       int
	lbStrategy       string
	lbDecision       string
}

var (
	backendTraceSubmitOnce    sync.Once
	backendTraceSubmitQ       chan *backendTraceJob
	backendTraceSubmitWorkers int
	backendTraceDropped       atomic.Uint64
	backendTraceInflight      atomic.Int64
)

// submitBackendTraceJob 把已快照的后端追踪交给固定 worker 组对象并写入。
// 满则短等 asyncq.EnqueueWait，仍满丢弃；计入同一实例 pending，关写入器时一起等。
func submitBackendTraceJob(job *backendTraceJob) {
	if job == nil || job.instanceID == "" || job.serviceID == "" || job.traceID == "" {
		return
	}
	ensureBackendTraceSubmit()
	accessLogPending.add(job.instanceID, 1)
	if asyncq.Offer(backendTraceSubmitQ, job) {
		return
	}
	accessLogPending.add(job.instanceID, -1)
	n := backendTraceDropped.Add(1)
	if n == 1 || n%1000 == 0 {
		logger.Warn("后端追踪提交队列已满，丢弃新日志",
			"dropped", n,
			"queueCap", cap(backendTraceSubmitQ),
			"queued", len(backendTraceSubmitQ),
			"inflight", backendTraceInflight.Load(),
			"instanceID", job.instanceID,
			"traceId", job.traceID)
	}
}

func ensureBackendTraceSubmit() {
	backendTraceSubmitOnce.Do(func() {
		backendTraceSubmitQ = make(chan *backendTraceJob, accessLogSubmitQueueSize)
		backendTraceSubmitWorkers = accessLogSubmitWorkerCount()
		for i := 0; i < backendTraceSubmitWorkers; i++ {
			go backendTraceSubmitWorker()
		}
		logger.Info("后端追踪提交队列已启动",
			"workers", backendTraceSubmitWorkers,
			"queueCap", accessLogSubmitQueueSize,
			"enqueueWait", asyncq.EnqueueWait)
	})
}

func backendTraceSubmitWorker() {
	for job := range backendTraceSubmitQ {
		runBackendTraceJob(job)
	}
}

func runBackendTraceJob(job *backendTraceJob) {
	backendTraceInflight.Add(1)
	defer backendTraceInflight.Add(-1)
	defer accessLogPending.add(job.instanceID, -1)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in backend trace writer",
				"error", r,
				"instanceID", job.instanceID,
				"traceId", job.traceID,
				"serviceId", job.serviceID)
		}
	}()
	writeBackendTraceFromJob(job)
}

func backendTraceSubmitIdle() bool {
	return backendTraceSubmitQ == nil ||
		(len(backendTraceSubmitQ) == 0 && backendTraceInflight.Load() == 0)
}
