package bees

import (
	"context"
	"fmt"
	"time"
)

func Example() {
	pool := Create(context.Background(), &Config{
		MaxWorkersCount: 1,
		IdleTimeout:     time.Minute,
		TimeoutJitter:   1,
	}, func(ctx context.Context, task interface{}) { fmt.Println(task) })
	defer pool.Close()

	pool.Submit(1)
	pool.Submit(2)
	pool.Submit(3)
	// Output:
	// 1
	// 2
	// 3
}
