# LocklessGenericRingBuffer

This is an implementation of a single producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18`. Instead of passing typeless `interface{}` which we have to assert or deserialized `[]byte`'s we are able to 
pass serialized structs between go routines in a type safe manner.

This package goes to great lengths not to allocate in the critical path and thus makes `0` allocations once the buffer is 
created outside the creation of consumers. 

A large part of the benefit of ring buffers can be attributed to the underlying array being continuous memory. 
Understanding how your structs lay out in memory 
([a brief introduction into how structs are represented in memory](https://research.swtch.com/godata)) is key to if your 
use case will benefit from storing the structs themselves vs pointers.

## Requirements
- `golang 1.18beta`

## Examples

### Create and Consume 
```go
var buffer = CreateBuffer[int](10)

messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
consumer, _ := buffer.CreateConsumer()

for _, value := range messages {
	buffer.Write(value)
}

for _, _ = range messages {
	_ = consumer.Get()
}
```

## Benchmarks 

### Comparison against channels 
```sql
BenchmarkConsumerSequentialReadWriteLarge-8            9         117031505 ns/op            1078 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteLarge-8            4         281360188 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerSequentialReadWriteMedium-8        1027           1170170 ns/op            1076 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteMedium-8         441           2742069 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerSequentialReadWriteSmall-8       103389             11536 ns/op            1076 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteSmall-8        43818             27350 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerConcurrentReadWriteLarge-8            2         649446354 ns/op        492003740 B/op        64 allocs/op
BenchmarkChannelsConcurrentReadWriteLarge-8            1        1315443125 ns/op        492001744 B/op        56 allocs/op
BenchmarkConsumerConcurrentReadWriteMedium-8         182           6631077 ns/op         4102696 B/op         37 allocs/op
BenchmarkChannelsConcurrentReadWriteMedium-8         100          13292200 ns/op         4102614 B/op         33 allocs/op
BenchmarkConsumerConcurrentReadWriteSmall-8        27282             43918 ns/op           26440 B/op         21 allocs/op
BenchmarkChannelsConcurrentReadWriteSmall-8        21544             55965 ns/op           26240 B/op         17 allocs/op
```

In sequential benchmarks it is about `2x` the read write speed of channels and in concurrent benchmarks where 
operations can block it is under `2x` faster than the channel implementation. 

## Testing 

Tests can currently be run via `go test`, to detect race conditions run with `go test -race`. As of the latest commit and 
with the current test it passes go's race condition detection however this does not mean it is race condition free. 
Additional tests, especially on creating and removing consumers in concurrent environments are needed. 

## TODO:
- [ ] formal benchmarks and performance tests
- [ ] Formal TLS Proof and Write up of algorithm
