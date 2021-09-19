# Bees

[![codecov](https://codecov.io/gh/delivery-club/bees/branch/master/graph/badge.svg)](https://codecov.io/gh/delivery-club/bees)
[![Go Report Card](https://goreportcard.com/badge/github.com/delivery-club/bees)](https://goreportcard.com/report/github.com/delivery-club/bees)

Bees - lightweight worker pool for go.

### Benchmarks:
![10m tasks and 500k workers](https://github.com/delivery-club/bees/blob/master/pool_bench_test.go):

#### WorkerPool:
<img width="900" alt="WorkerPoolBench" src="https://user-images.githubusercontent.com/27820873/133930212-806c5918-4b30-4950-8139-326317ce3a56.png">

#### Gorutines:
<img width="900" alt="GoroutinesBench" src="https://user-images.githubusercontent.com/27820873/133930166-d34b6dcf-b9f0-4275-93ec-08ecdb988e1f.png">

#### Semaphore:
<img width="900" alt="SemaphoreBench" src="https://user-images.githubusercontent.com/27820873/133930179-25495409-65cb-447a-ab06-72698412c646.png">
