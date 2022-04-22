# LocklessGenericRingBuffer

This is an implementation of a single producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18`. Instead of passing typeless `interface{}` which we have to assert or deserialized `[]byte`'s we are able to 
pass serialized structs between go routines in a type safe manner. 

**Note: the current implementation writers nor individual consumers is NOT thread safe, i.e. one consumer per go routine** 


This package goes to great lengths not to allocate in the critical path and thus makes `0` allocations once the buffer is 
created outside the creation of consumers. 

A large part of the benefit of ring buffers can be attributed to the underlying array being continuous memory. 
Understanding how your structs lay out in memory 
([a brief introduction into how structs are represented in memory](https://research.swtch.com/godata)) is key to if your 
use case will benefit from storing the structs themselves vs pointers to your desired type.

## Requirements
- `golang 1.18.x or above`

## Examples

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

Benchmarks are ran on a **M1 Macbook Air (16gb ram)**, each benchmark does not include creation time of the consumers/channels. 
This **favours** the channels' implementation compared to benchmark's including the creation time , however it is a 
better representation of the use case for a lockless ring buffer.


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

In sequential benchmarks it is about `2x` the read write speed of channels and in concurrent benchmarks, where 
operations can block, it is slightly `2x` faster than the channel implementation on a large amount of messages **1M+** and slightly
under `2x` with a small amount i.e processing sub 100. 

## Testing 

To run current tests: `go test`

To detect race conditions run with `go test -race`, which as of the latest commit (January 24, 2021) with the current test cases it 
passes. 

**Note** this does not mean it is race condition free. 
Additional tests, especially on creating and removing consumers in concurrent environments are needed. 

## TODO:
- [ ] formal benchmarks and performance tests
- [ ] Formal TLS Proof and Write up of algorithm
