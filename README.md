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
BenchmarkConsumerSequentialReadWrite-8               849           1261116 ns/op            1060 B/op          4 allocs/op
BenchmarkChannelsSequentialReadWrite-8               453           2640028 ns/op             896 B/op          1 allocs/op
BenchmarkConsumerConcurrentReadWrite-8               235           5095339 ns/op         4102668 B/op         37 allocs/op
BenchmarkChannelsConcurrentReadWrite-8                79          15557314 ns/op         4102493 B/op         33 allocs/op
```

In sequential benchmarks it is about double the read write speed of channels and in concurrent benchmarks where 
operations can block it is over three times faster than the channel implementation. 

## TODO:
- [ ] formal benchmarks and performance tests
- [ ] Formal TLS Proof and Write up of algorithm
