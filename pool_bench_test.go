package bees

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	_ = 1 << (10 * iota)
	_
	MiB
)

const (
	poolSize = 500000
	sleep    = 10
	runTimes = 10000000
)

func BenchmarkSemaphore(b *testing.B) {
	var wg sync.WaitGroup
	sema := make(chan struct{}, poolSize)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoFunc()
				<-sema
				wg.Done()
			}()
		}
	}
	wg.Wait()
	b.StopTimer()

	b.Logf("memory usage:%d MB", checkMem())
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
	}
	wg.Wait()
	b.StopTimer()

	b.Logf("memory usage:%d MB", checkMem())
}

func BenchmarkWorkerPool(b *testing.B) {
	var wg sync.WaitGroup

	p := Create(context.Background(), &Config{
		MaxWorkersCount: poolSize,
		IdleTimeout:     5 * time.Second,
		TimeoutJitter:   0,
	}, func(ctx context.Context, task interface{}) {
		demoFunc()
		wg.Done()
	})
	defer func() {
		p.Close()
	}()
	var task interface{}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			p.Submit(task)
		}
	}
	wg.Wait()
	b.StopTimer()

	b.Logf("memory usage:%d MB", checkMem())
}

func checkMem() uint64 {
	var curMem uint64
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	return curMem
}

func demoFunc() {
	time.Sleep(time.Duration(sleep) * time.Millisecond)
}
