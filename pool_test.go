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

	pool := Create(context.Background(), func(ctx context.Context, i interface{}) {},
		WithCapacity(3), WithKeepAlive(50*time.Second), WithJitter(1))
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

	pool := Create(context.Background(), func(ctx context.Context, i interface{}) {
		select {
		case <-time.After(time.Hour):
		case <-ctx.Done():
		}
	}, WithCapacity(1), WithKeepAlive(time.Hour))

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

	pool := Create(context.Background(), func(ctx context.Context, i interface{}) {},
		WithKeepAlive(0), WithJitter(0), WithCapacity(0))

	if pool.cfg.KeepAliveTimeout == 0 || pool.cfg.TimeoutJitter == 0 || pool.cfg.Capacity == 0 {
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
	pool := Create(context.Background(), func(ctx context.Context, i interface{}) {
		defer wg.Done()
		panic("aaaaaa")
	}, WithKeepAlive(time.Hour), WithJitter(10), WithCapacity(1))

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

	pool := Create(context.Background(), func(ctx context.Context, i interface{}) {},
		WithKeepAlive(10*time.Millisecond), WithJitter(1), WithCapacity(1),
	)

	for i := 0; i < 100; i++ {
		pool.Submit(task)
	}
	time.Sleep(100 * time.Millisecond)

	if *pool.activeWorkers != 0 || *pool.freeWorkers != 0 {
		t.Fatalf("active workers found")
	}
}
