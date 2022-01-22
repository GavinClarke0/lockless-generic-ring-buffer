package lockless_generic_ring_buffer

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	MaxConsumerError = errors.New("max amount of consumers reached cannot create any more")
)

type RingBuffer[T any] struct {
	buffer                []T
	length                uint32
	headPointer           uint32 // next position to write
	minTailPointerCounter []uint32
	minTailPointer        uint32
	maxReaderIndex        uint32
	readerCount           uint32
	readerPointers        []*uint32
	consumerLock          sync.Mutex
}

type Consumer[T any] struct {
	ring *RingBuffer[T]
	id   uint32
}

func CreateBuffer[T any](size uint32, maxConsumer uint32) RingBuffer[T] {

	return RingBuffer[T]{
		buffer:                make([]T, size, size),
		minTailPointerCounter: make([]uint32, size, size),
		length:                size,
		headPointer:           0,
		maxReaderIndex:        0,
		minTailPointer:        ^uint32(0),
		readerPointers:        make([]*uint32, maxConsumer+1),
		consumerLock:          sync.Mutex{},
	}
}

/*
CreateConsumer

Create a consumer by assigning it the id of the first empty position in the consumerPosition array. A nil value represents
an unclaimed/not used consumer.

Locks can be used as it has no effect on read/write operations and is only to keep consumer consistency, thus the
algorithm is still lockless For best performance, consumers should be preallocated before starting buffer operations
*/
func (ringbuffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	ringbuffer.consumerLock.Lock()
	defer ringbuffer.consumerLock.Unlock()

	var insertIndex = ringbuffer.length // default to maximum case

	// locate first empty position, if no empty positions return error
	for i, consumer := range ringbuffer.readerPointers {
		if consumer == nil {
			insertIndex = uint32(i)
			break
		}
	}

	if int(insertIndex) == len(ringbuffer.readerPointers) {
		return Consumer[T]{}, MaxConsumerError
	}

	if insertIndex >= ringbuffer.maxReaderIndex {
		atomic.AddUint32(&ringbuffer.maxReaderIndex, 1)
	}

	var readPosition = ringbuffer.headPointer - 1
	ringbuffer.readerPointers[insertIndex] = &readPosition

	atomic.AddUint32(&ringbuffer.readerCount, 1)

	return Consumer[T]{
		id:   insertIndex,
		ring: ringbuffer,
	}, nil
}

func (consumer *Consumer[T]) Remove() {
	consumer.ring.removeConsumer(consumer.id)
}

func (consumer *Consumer[T]) Get() T {
	return consumer.ring.readIndex(consumer.id)
}

func (ringbuffer *RingBuffer[T]) removeConsumer(consumerId uint32) {

	ringbuffer.consumerLock.Lock()
	defer ringbuffer.consumerLock.Unlock()

	ringbuffer.readerPointers[consumerId] = nil
	atomic.CompareAndSwapUint32(&ringbuffer.maxReaderIndex, consumerId-1, ringbuffer.maxReaderIndex-1)
}

func (ringbuffer *RingBuffer[T]) Write(value T) {

	var lastTailReaderPointerPosition uint32
	//var currentReadPosition uint32
	// var currentMaxReaderPosition uint32
	// var i uint32

	/*
		We are blocking until the all at least one space is available in the buffer to write.

		As overflow properties of uint32 are utilized to ensure slice index boundaries are adhered too we add length of
		buffer to current consumer read positions allowing us to determine the least read consumer.

		For example: buffer of size 2

		uint8 head = 1
		uint8 tail = 255
		tail + 2 => 1 with overflow, same as buffer
	*/
	for {
		lastTailReaderPointerPosition = atomic.LoadUint32(&ringbuffer.minTailPointer) + ringbuffer.length

		if lastTailReaderPointerPosition > ringbuffer.headPointer {
			ringbuffer.buffer[ringbuffer.headPointer%ringbuffer.length] = value

			atomic.StoreUint32(&ringbuffer.minTailPointerCounter[ringbuffer.headPointer%ringbuffer.length], atomic.LoadUint32(&ringbuffer.readerCount))
			atomic.AddUint32(&ringbuffer.headPointer, 1)

			return
		}
		runtime.Gosched()
	}
}

func (ringbuffer *RingBuffer[T]) readIndex(consumerId uint32) T {

	var newIndex = atomic.AddUint32(ringbuffer.readerPointers[consumerId], 1)

	// yield until work is available
	for newIndex >= atomic.LoadUint32(&ringbuffer.headPointer) {
		runtime.Gosched()
	}

	value := ringbuffer.buffer[newIndex%ringbuffer.length]
	if atomic.AddUint32(&ringbuffer.minTailPointerCounter[newIndex%ringbuffer.length], ^uint32(0)) == 0 {
		atomic.AddUint32(&ringbuffer.minTailPointer, 1)
	}

	return value
}
