package ringbuffer

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	MaxUint32        = ^uint32(0)
	MaxConsumers     = 200
	MaxConsumerError = errors.New("max amount of consumers reached cannot create any more")
)

type RingBuffer[T any] struct {
	buffer          []T
	length          uint32
	headPointer     uint32 // next position to write
	maxReaderIndex  uint32
	readersPosition [200]*uint32
	consumerLock    sync.Mutex
}

type Consumer[T any] struct {
	ring *RingBuffer[T]
	id   uint32
}

func CreateBuffer[T any](size uint32) RingBuffer[T] {

	return RingBuffer[T]{
		buffer:          make([]T, size, size),
		length:          size,
		headPointer:     0,
		readersPosition: [200]*uint32{},
		consumerLock:    sync.Mutex{},
	}
}

/*
CreateConsumer

Create a consumer by assiging it the id of the first empty position in the consumerPosition array.A nil value represents
a unclaimed/not used consumer if

Locks can be used as it has no effect on read/write operations and is only to keep consumer consistancy, thus the
alogrithmn.is still lockless For best preformance,consumers should be preallocated before starting buffer operations
*/
func (ringbuffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	ringbuffer.consumerLock.Lock()
	defer ringbuffer.consumerLock.Unlock()

	var insertIndex = ringbuffer.length // default to maximum case

	// locate first empty position, if no empty positions return error
	for i, consumer := range ringbuffer.readersPosition {
		if consumer == nil {
			insertIndex = uint32(i)
			break
		}
	}

	if insertIndex == ringbuffer.length {
		return Consumer[T]{}, MaxConsumerError
	}

	if insertIndex >= ringbuffer.maxReaderIndex {
		atomic.AddUint32(&ringbuffer.maxReaderIndex, 1)
	}

	var readPosition = ringbuffer.headPointer - 1
	ringbuffer.readersPosition[insertIndex] = &readPosition

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

	ringbuffer.readersPosition[consumerId] = nil

	if consumerId == ringbuffer.maxReaderIndex-1 {
		ringbuffer.maxReaderIndex--
	}
}

func (ringbuffer *RingBuffer[T]) Write(value T) {

	var lastRead uint32
	var currentRead uint32
	var i uint32

	/*
		Non-critical path, we are blocking until the all at least one space is available in the buffer to write.

		To allow overflow wrap of uint32 value we add length of buffer to current consumer read positions allowing us to
		determine the least read consumer at the overflow boundary.

		For example: buffer of size 2

		uint8 head = 1
		uint8 tail = 255
		tail + 2 => 1 with overflow, same as buffer

		Concurrent access does not matter as variables can only be updated to make progression towards head pointer, if a
		variable is updated during iteration it has no effect on data integrity. Worst case it causes another yield to
		scheduler and recheck for current minimum consumer.
	*/
	for {
		lastRead = ringbuffer.headPointer + ringbuffer.length
		for i = 0; i <= ringbuffer.maxReaderIndex; i++ {

			if ringbuffer.readersPosition[i] == nil {
				continue
			}

			currentRead = *ringbuffer.readersPosition[i] + ringbuffer.length
			if currentRead < lastRead {
				lastRead = currentRead
			}
		}

		if lastRead > ringbuffer.headPointer {
			break
		}
		runtime.Gosched()
	}

	ringbuffer.buffer[ringbuffer.headPointer%ringbuffer.length] = value
	atomic.AddUint32(&ringbuffer.headPointer, 1)
}

func (ringbuffer *RingBuffer[T]) readIndex(consumerId uint32) T {

	var newIndex = atomic.AddUint32(ringbuffer.readersPosition[consumerId], 1)

	// yield until work is available
	for newIndex >= ringbuffer.headPointer {
		runtime.Gosched()
	}
	return ringbuffer.buffer[newIndex%ringbuffer.length]
}
