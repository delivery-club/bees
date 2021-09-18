package bees

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func TestPoolSpawnWorker(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	pool := Create(ctx, &Config{
		MaxWorkersCount: 3,
		IdleTimeout:     50 * time.Second,
		TimeoutJitter:   1,
	}, func(ctx context.Context, i interface{}) {})
	pool.SetLogger(log.Default())

	var wg sync.WaitGroup

	count := 10
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			pool.retrieveWorker()
			wg.Done()
		}()
	}
	wg.Wait()

	cancel()
	pool.Close()
	time.Sleep(100 * time.Millisecond)

	if *pool.activeWorkers != 0 {
		t.Fatalf("active workers not zero")
	}
}

func TestPoolShutdownWithStackedWorker(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	pool := Create(ctx, &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Hour,
	}, func(ctx context.Context, i interface{}) {
		select {
		case <-time.After(time.Hour):
		case <-ctx.Done():
		}
	})

	var task interface{}

	pool.Submit(task)

	var wg sync.WaitGroup

	// заполняем пул заданий
	pool.Submit(task)

	// создаем пробку
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { pool.SubmitAsync(task); wg.Done() }()
	}

	// ждем расхождения пробки
	cancel()
	wg.Wait()

	pool.Close()
}

func TestPoolDefaultValues(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("uncatched panic")
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
			t.Fatalf("uncatched panic")
		}
	}()

	var wg sync.WaitGroup
	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Hour,
		TimeoutJitter:   10,
	}, func(ctx context.Context, i interface{}) { wg.Done(); panic("aaaaaa") })

	var task interface{}

	wg.Add(1)
	pool.Submit(task)
	wg.Wait()
}

func TestPoolExpiration(t *testing.T) {
	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Nanosecond,
		TimeoutJitter:   1,
	}, func(ctx context.Context, i interface{}) {})

	pool.Submit(1)
	pool.wg.Wait()
}
