# Bees

[![codecov](https://codecov.io/gh/delivery-club/bees/branch/master/graph/badge.svg)](https://codecov.io/gh/delivery-club/bees)
[![Go Report Card](https://goreportcard.com/badge/github.com/delivery-club/bees)](https://goreportcard.com/report/github.com/delivery-club/bees)
[![Go Reference](https://pkg.go.dev/badge/github.com/delivery-club/bees.svg)](https://pkg.go.dev/github.com/delivery-club/bees)

Bees - simple and lightweight worker pool for go.

## Benchmarks:

### [10m tasks and 500k workers](https://github.com/delivery-club/bees/blob/master/pool_bench_test.go):

#### WorkerPool:

<img width="900" alt="WorkerPoolBench" src="https://user-images.githubusercontent.com/27820873/133930212-806c5918-4b30-4950-8139-326317ce3a56.png">

<b>only 37MB used for 500k workers pool</b>

#### Gorutines:

<img width="900" alt="GoroutinesBench" src="https://user-images.githubusercontent.com/27820873/133930166-d34b6dcf-b9f0-4275-93ec-08ecdb988e1f.png">

#### Semaphore:

<img width="900" alt="SemaphoreBench" src="https://user-images.githubusercontent.com/27820873/133930179-25495409-65cb-447a-ab06-72698412c646.png">

## Examples:

```go
// Example - demonstrate pool usage
func Example() {
    pool := Create(context.Background(),
        func(ctx context.Context, task interface{}) { fmt.Println(task) },
    )

    defer pool.Close()

    pool.Submit(1)
    pool.Submit(2)
    pool.Submit(3)

    // Output:
    // 1
    // 2
    // 3
}
```
