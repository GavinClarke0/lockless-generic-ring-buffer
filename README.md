# LocklessGenericRingBuffer

This is an implementation of a single producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18`. Instead of passing typeless `interface{}` which we have to assert or deserialized `[]byte`'s we are able to 
pass serialized structs between go routines in a type safe manner.

This package goes to great lengths not to allocate in the critical path and thus makes 0 allocations once the buffer is 
created outside the creation of consumers. 

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
BenchmarkConsumerSequentialReadWriteLarge-8            8         127404589 ns/op            1060 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteLarge-8            4         265133938 ns/op             898 B/op          1 allocs/op
BenchmarkConsumerSequentialReadWriteMedium-8         942           1264458 ns/op            1060 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteMedium-8         452           2655275 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerSequentialReadWriteSmall-8        94333             12593 ns/op            1060 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWriteSmall-8        45060             26648 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerConcurrentReadWriteLarge-8            2         520416396 ns/op        492003812 B/op        65 allocs/op
BenchmarkChannelsConcurrentReadWriteLarge-8            1        1245517208 ns/op        492002264 B/op        59 allocs/op
BenchmarkConsumerConcurrentReadWriteMedium-8         235           5097063 ns/op         4102692 B/op         37 allocs/op
BenchmarkChannelsConcurrentReadWriteMedium-8         100          12909553 ns/op         4102507 B/op         33 allocs/op
BenchmarkConsumerConcurrentReadWriteSmall-8        30693             39371 ns/op           26424 B/op         21 allocs/op
BenchmarkChannelsConcurrentReadWriteSmall-8        22174             53359 ns/op           26241 B/op         17 allocs/op
```

In sequential benchmarks it is about double the read write speed of channels and in concurrent benchmarks where 
operations can block it is over three times faster than the channel implementation. 

## TODO:
- [ ] formal benchmarks and performance tests
- [ ] Formal TLS Proof and Write up of algorithm
