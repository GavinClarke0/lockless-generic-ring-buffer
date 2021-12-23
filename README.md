## LocklessGenericRingBuffer

This is an implementation of a multi producer, multi reader lockless ring buffer utilizing the new generics available in 
`go 1.18`. Instead of passing typeless `interface{}` which we have to assert or deserialized `[]byte` we are able to 
pass serialized structs between go routines in a type safe manner.

### Requirements
- `golang 1.18beta`
- 