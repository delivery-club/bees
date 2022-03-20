package bees

import (
	"context"
	"fmt"
	"time"
)

// Example - demonstrate pool usage
func Example() {
	pool := Create(context.Background(), func(ctx context.Context, task interface{}) { fmt.Println(task) },
		WithCapacity(1), WithKeepAlive(time.Minute), WithoutJitter,
	)
	defer pool.Close()

	pool.Submit(1)
	pool.Submit(2)
	pool.Submit(3)
	pool.Wait()
	// Output:
	// 1
	// 2
	// 3
}
