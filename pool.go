package bees

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Config - configuration struct for WorkerPool used only in Create method
type Config struct {
	MaxWorkersCount int64
	IdleTimeout     time.Duration
	TimeoutJitter   int // additional random timeout in ms
}

// WorkerPool - keep information about active/free workers and task processing
type WorkerPool struct {
	activeWorkers   *int64
	freeWorkers     *int64
	workersCapacity int64

	process TaskProcessor
	taskCh  chan interface{}

	cfg         *Config
	shutdownCtx context.Context
	cancelFunc  context.CancelFunc
	wg          *sync.WaitGroup

	logger logger
}

type logger interface {
	Printf(format string, args ...interface{})
}

// TaskProcessor - closure type for processing tasks
type TaskProcessor func(ctx context.Context, task interface{})

// Create - create worker pool instance
func Create(ctx context.Context, config *Config, processor TaskProcessor) *WorkerPool {
	if config.TimeoutJitter <= 0 {
		config.TimeoutJitter = 1
	}

	if config.IdleTimeout <= 0 {
		config.IdleTimeout = time.Second
	}

	if config.MaxWorkersCount <= 0 {
		config.MaxWorkersCount = 1
	}
	ctx, cancel := context.WithCancel(ctx)

	return &WorkerPool{
		activeWorkers:   ptrOfInt64(0),
		freeWorkers:     ptrOfInt64(0),
		workersCapacity: config.MaxWorkersCount,
		process:         processor,
		taskCh:          make(chan interface{}, 2*config.MaxWorkersCount),
		cfg:             config,
		shutdownCtx:     ctx,
		cancelFunc:      cancel,
		wg:              &sync.WaitGroup{},
		logger:          log.Default(),
	}
}

// SetLogger - sets logger for pool
func (wp *WorkerPool) SetLogger(logger logger) {
	wp.logger = logger
}

// Submit - submit task to pool
func (wp *WorkerPool) Submit(task interface{}) {
	wp.retrieveWorker()
	wp.taskCh <- task
}

// SubmitAsync - submit task to pool, may call async
func (wp *WorkerPool) SubmitAsync(task interface{}) {
	wp.retrieveWorker()
	select {
	case wp.taskCh <- task:
	case <-wp.shutdownCtx.Done():
	}
}

func (wp *WorkerPool) retrieveWorker() {
	if free := atomic.LoadInt64(wp.freeWorkers); free > 1 {
		return
	}

	if c := atomic.LoadInt64(wp.activeWorkers); c < wp.workersCapacity {
		if atomic.CompareAndSwapInt64(wp.activeWorkers, c, c+1) {
			wp.spawnWorker()
		}
	}
}

func (wp *WorkerPool) spawnWorker() {
	atomic.AddInt64(wp.freeWorkers, 1)
	wp.wg.Add(1)

	go func() {
		// https://en.wikipedia.org/wiki/Exponential_backoff
		// nolint:gosec
		jitter := time.Millisecond * time.Duration(rand.Intn(wp.cfg.TimeoutJitter))
		ticker := time.NewTicker(wp.cfg.IdleTimeout + jitter)
		defer ticker.Stop()

		defer func() {
			atomic.AddInt64(wp.freeWorkers, -1)
			atomic.AddInt64(wp.activeWorkers, -1)
			wp.wg.Done()
			if err := recover(); err != nil {
				atomic.AddInt64(wp.freeWorkers, 1)

				if wp.workersCapacity == 1 {
					go wp.retrieveWorker()
				}

				wp.logger.Printf("on WorkerPool: on Process: %+v", err)
				return
			}
		}()

		for {
			select {
			case task := <-wp.taskCh:
				atomic.AddInt64(wp.freeWorkers, -1)
				wp.process(wp.shutdownCtx, task)
				atomic.AddInt64(wp.freeWorkers, 1)
			case <-wp.shutdownCtx.Done():
				return
			case <-ticker.C:
				return
			}
			ticker.Reset(wp.cfg.IdleTimeout)
		}
	}()
}

// Close - close worker pool and release all resources, not processed tasks will be throw away
func (wp *WorkerPool) Close() {
	wp.cancelFunc()
	wp.wg.Wait()
}

func ptrOfInt64(i int64) *int64 {
	return &i
}
