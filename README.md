## LocklessGenericRingBuffer

This is an implementation of a multi producer, multi consumer lockless ring buffer utilizing the new generics available in go 1.18. Instead of passing 
byte arrays as mesages we are able to pass deserialized structs between go routine/publishers susbcribers in a type safe manner.


### Requirments 

- `golang 1.18beta`