# LocklessGenericRingBuffer

This is an implementation of a single producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18`. Instead of passing typeless `interface{}` which we have to assert or deserialized `[]byte`'s we are able to 
pass serialized structs between go routines in a type safe manner.

This package goes to great lengths not to allocate in the critical path and thus makes 0 allocations once the buffer is created. 

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

## TODO:
- [ ] formal benchmarks and preformance tests
- [ ] Formal TLS Proof and Write up of algorithm
