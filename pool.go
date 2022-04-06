package bees

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool - keep information about active/free workers and task processing
type WorkerPool struct {
	activeWorkers   *int64
	freeWorkers     *int64
	taskCount       *int64
	workersCapacity *int64

	process TaskProcessor
	taskCh  chan interface{}

	cfg         *config
	shutdownCtx context.Context
	cancelFunc  context.CancelFunc
	wg          *sync.WaitGroup
	isClosed    *int64

	logger logger
}

// Create - create worker pool instance
func Create(ctx context.Context, processor TaskProcessor, opts ...Option) *WorkerPool {
	var config = config{
		Capacity:         1,
		TimeoutJitter:    1000,
		KeepAliveTimeout: time.Minute,
		GracefulTimeout:  time.Minute,
	}

	for _, opt := range opts {
		opt.apply(&config)
	}

	if config.TimeoutJitter <= 0 {
		config.TimeoutJitter = 1
	}

	if config.KeepAliveTimeout <= 0 {
		config.KeepAliveTimeout = time.Second
	}

	if config.Capacity <= 0 {
		config.Capacity = 1
	}
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	return &WorkerPool{
		activeWorkers:   ptrOfInt64(0),
		freeWorkers:     ptrOfInt64(0),
		taskCount:       ptrOfInt64(0),
		workersCapacity: ptrOfInt64(config.Capacity),
		process:         processor,
		taskCh:          make(chan interface{}, 2*config.Capacity),
		cfg:             &config,
		shutdownCtx:     ctx,
		cancelFunc:      cancel,
		wg:              wg,
		logger:          log.Default(),
		isClosed:        ptrOfInt64(0),
	}
}

// SetLogger - sets logger for pool
func (wp *WorkerPool) SetLogger(logger logger) {
	wp.logger = logger
}

// Submit - submit task to pool
func (wp *WorkerPool) Submit(task interface{}) {
	if atomic.LoadInt64(wp.isClosed) == 1 {
		return
	}

	wp.retrieveWorker()
	wp.taskCh <- task
	atomic.AddInt64(wp.taskCount, 1)
}

// SubmitAsync - submit task to pool, for async better use this method
func (wp *WorkerPool) SubmitAsync(task interface{}) {
	if atomic.LoadInt64(wp.isClosed) == 1 {
		return
	}

	wp.retrieveWorker()
	select {
	case wp.taskCh <- task:
		atomic.AddInt64(wp.taskCount, 1)
	case <-wp.shutdownCtx.Done():
	}
}

func (wp *WorkerPool) Wait() {
	const maxBackoff = 16
	backoff := 1

	for atomic.LoadInt64(wp.taskCount) != 0 {
		for i := 0; i < backoff; i++ {
			time.Sleep(time.Duration(backoff) * time.Microsecond)
		}
		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

// Close - close worker pool and release all resources, not processed tasks will be thrown away
func (wp *WorkerPool) Close() {
	atomic.StoreInt64(wp.isClosed, 1)
	wp.cancelFunc()
	wp.wg.Add(-1)
	wp.wg.Wait()
}

// CloseGracefully - close worker pool and release all resources, wait until all task will be processed
func (wp *WorkerPool) CloseGracefully() {
	atomic.StoreInt64(wp.isClosed, 1)
	closed := make(chan struct{})
	go func() {
		wp.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(wp.cfg.GracefulTimeout):
	}

	wp.cancelFunc()

	wp.wg.Add(-1)
	wp.wg.Wait()
}

func (wp *WorkerPool) Scale(delta int64) {
	atomic.AddInt64(wp.workersCapacity, delta)
}

func (wp *WorkerPool) retrieveWorker() {
	max := atomic.LoadInt64(wp.workersCapacity)
	c := atomic.LoadInt64(wp.activeWorkers)

	if c < max {
		if atomic.CompareAndSwapInt64(wp.activeWorkers, c, c+1) {
			wp.spawnWorker()
		}
	}
}

func (wp *WorkerPool) spawnWorker() {
	atomic.AddInt64(wp.freeWorkers, 1)
	wp.wg.Add(1)

	go func() {
		randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
		// https://en.wikipedia.org/wiki/Exponential_backoff
		// nolint:gosec
		ticker := time.NewTicker(wp.cfg.KeepAliveTimeout + time.Millisecond*time.Duration(randomizer.Intn(wp.cfg.TimeoutJitter)))
		defer ticker.Stop()

		defer func() {
			atomic.AddInt64(wp.freeWorkers, -1)
			atomic.AddInt64(wp.activeWorkers, -1)
			wp.wg.Done()
			if err := recover(); err != nil {
				atomic.AddInt64(wp.freeWorkers, 1)
				atomic.AddInt64(wp.taskCount, -1)
				if atomic.LoadInt64(wp.activeWorkers) == 0 {
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
				atomic.AddInt64(wp.taskCount, -1)
			case <-wp.shutdownCtx.Done():
				return
			case <-ticker.C:
				return
			}
			ticker.Reset(wp.cfg.KeepAliveTimeout + time.Millisecond*time.Duration(randomizer.Intn(wp.cfg.TimeoutJitter)))
		}
	}()
}

func ptrOfInt64(i int64) *int64 {
	return &i
}
