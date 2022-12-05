# LocklessGenericRingBuffer

LocklessGenericRingBuffer is a single producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18+`. Instead of passing typeless `interface{}` or byte arrays we are able to pass serialized structs between go routines in a type safe manner. 

### What is a lockless ringbuffer ?

A ring buffer, also known as a circular buffer, is a fixed-size buffer that can be efficiently appended to and read from. This implementation allows for multiple goroutines to concurrently read and a single goroutine to write to the buffer without the need for locks, ensuring maximum throughput and minimal latency. 


## Benefits

- [x] Zero Heap Allocations
- [x] Cache Friendly are underlying structures are continuous memory
- [x] Faster then channels for highly contended workloads (See Benchmarks)
- [x] Zero Dependencies 

## Requirements
- `golang 1.18.x or above`

## Getting started

**Note: writers and consumers are NOT thread safe, i.e. only use a consumer in a single go routine** 

### Install 

```
go get github.com/GavinClarke0/lockless-generic-ring-buffer
```

### Create and Consume 
```go
var buffer, _ = CreateBuffer[int](16) // buffer size must be to the power 2

messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
consumer, _ := buffer.CreateConsumer()

for _, value := range messages {
	buffer.Write(value)
}

for _, _ = range messages {
	_ = consumer.Get()
}
```

### Remove a Consumer
```go
var consumer, _ = buffer.CreateConsumer()
consumer.Remove()
```

## Benchmarks 

### Comparison against channels 

Benchmarks are ran on a **M1 Macbook Air (16gb ram)**.

**Note: each benchmark does not include creation time of the consumers/channels.**

```sql
BenchmarkConsumerSequentialReadWriteLarge-8           20          55602675 ns/op               0 B/op          0 allocs/op
BenchmarkChannelsSequentialReadWriteLarge-8            8         133155344 ns/op               0 B/op          0 allocs/op
BenchmarkConsumerSequentialReadWriteMedium-8        1063           1123298 ns/op               0 B/op          0 allocs/op
BenchmarkChannelsSequentialReadWriteMedium-8         451           2650842 ns/op               0 B/op          0 allocs/op
BenchmarkConsumerSequentialReadWriteSmall-8        99393             12099 ns/op               0 B/op          0 allocs/op
BenchmarkChannelsSequentialReadWriteSmall-8        41755             28758 ns/op               0 B/op          0 allocs/op
BenchmarkConsumerConcurrentReadWriteLarge-8            5         223985800 ns/op             345 B/op          2 allocs/op
BenchmarkChannelsConcurrentReadWriteLarge-8            2         858931292 ns/op             144 B/op          2 allocs/op
BenchmarkConsumerConcurrentReadWriteMedium-8         278           4554057 ns/op             217 B/op          2 allocs/op
BenchmarkChannelsConcurrentReadWriteMedium-8          90          17578294 ns/op             169 B/op          2 allocs/op
BenchmarkConsumerConcurrentReadWriteSmall-8        36378             33837 ns/op              96 B/op          2 allocs/op
BenchmarkChannelsConcurrentReadWriteSmall-8        25004             47466 ns/op              97 B/op          2 allocs/op

```

In sequential benchmarks it is about `2x` the read write speed of channels. 

In concurrent benchmarks, where operations can block, it is about `2x` faster than the channel implementation. 

## Testing 

To run current tests: `go test`

To detect race conditions run with `go test -race`, which as of the latest commit (January 24, 2021) with the current test cases it 
passes. 

**Note** this does not mean it is race condition free. 
Additional tests, especially on creating and removing consumers in concurrent environments are needed. 
