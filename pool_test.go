package bees

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

const task = 1

func TestWorkerPool_Close(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic: %+v", err)
		}
	}()

	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 3,
		IdleTimeout:     50 * time.Second,
		TimeoutJitter:   1,
	}, func(ctx context.Context, i interface{}) {})
	pool.SetLogger(log.Default())

	for i := 0; i < 100; i++ {
		go pool.Submit(task)
		go pool.SubmitAsync(task)
	}

	pool.Close()
}

func TestWorkerPool_ShutdownWithStacked(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic: %+v", err)
		}
	}()

	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Hour,
	}, func(ctx context.Context, i interface{}) {
		select {
		case <-time.After(time.Hour):
		case <-ctx.Done():
		}
	})

	var wg sync.WaitGroup
	// fill up worker pool by tasks
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.SubmitAsync(task)
		}()
	}

	pool.Close()
	wg.Wait()
}

func TestPoolDefaultValues(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic: %+v", err)
		}
	}()

	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 0,
		IdleTimeout:     0,
		TimeoutJitter:   0,
	}, func(ctx context.Context, i interface{}) {})

	if pool.cfg.IdleTimeout == 0 || pool.cfg.TimeoutJitter == 0 || pool.cfg.MaxWorkersCount == 0 {
		t.Fail()
	}
}

func TestPoolRecoverAfterPanicOnSingleWorker(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic: %+v", err)
		}
	}()

	var wg sync.WaitGroup
	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Hour,
		TimeoutJitter:   10,
	}, func(ctx context.Context, i interface{}) {
		defer wg.Done()
		panic("aaaaaa")
	})

	wg.Add(1)
	pool.Submit(task)
	wg.Wait()
}

func TestWorkerExpiration(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic: %+v", err)
		}
	}()

	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     10 * time.Millisecond,
		TimeoutJitter:   1,
	}, func(ctx context.Context, i interface{}) {})

	for i := 0; i < 100; i++ {
		pool.Submit(task)
	}
	time.Sleep(100 * time.Millisecond)

	if *pool.activeWorkers != 0 || *pool.freeWorkers != 0 {
		t.Fatalf("active workers found")
	}
}
