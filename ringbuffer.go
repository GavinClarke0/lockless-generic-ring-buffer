package ringbuffer

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	MaxUint32        = ^uint32(0)
	MaxConsumers     = 100
	MaxConsumerError = errors.New("max amount of consumers reached cannot create any more")
)

type RingBuffer[T any] struct {
	buffer          []T
	length          uint32
	headPointer     uint32 // next position to write
	readerCount     uint32
	readersPosition [100]*uint32
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
		readersPosition: [100]*uint32{},
		consumerLock:    sync.Mutex{},
	}
}

/*
Create a consumer by assiging it the id of the first empty position in the consumer arrays.

Locks can be used as it has no effect on read/write operations of locks and only to keep consumer consistancy.
For best preformance,consumers should be preallocated before starting buffer operations

*/
func (ringbuffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	ringbuffer.consumerLock.Lock()
	defer ringbuffer.consumerLock.Unlock()

	var insertIndex = ringbuffer.length

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

	// increment reader, since location is currently nil will be ignored until reader position set
	if insertIndex >= ringbuffer.readerCount {
		atomic.AddUint32(&ringbuffer.readerCount, 1)
	}

	var readPosition = ringbuffer.headPointer - 1
	ringbuffer.readersPosition[insertIndex] = &readPosition

	return Consumer[T]{
		id:   insertIndex,
		ring: ringbuffer,
	}, nil
}

func (ringbuffer *RingBuffer[T]) removeConsumer(consumerId uint32) {

	ringbuffer.consumerLock.Lock()
	defer ringbuffer.consumerLock.Unlock()

	ringbuffer.readersPosition[consumerId] = nil

	if consumerId == ringbuffer.readerCount-1 {
		ringbuffer.readerCount--
	}
}

func (consumer *Consumer[T]) Remove() {
	consumer.ring.removeConsumer(consumer.id)
}

func (ringbuffer *RingBuffer[T]) Write(value T) {

	var lastRead uint32
	var currentRead uint32
	var i uint32

	// Non-critical path, we are blocking until the all at least one space is available in the buffer to write
	lastRead = ringbuffer.headPointer + ringbuffer.length
	for i = 0; i <= ringbuffer.readerCount; i++ {

		if ringbuffer.readersPosition[i] == nil {
			continue
		}

		currentRead = *ringbuffer.readersPosition[i] + ringbuffer.length
		if currentRead < lastRead {
			lastRead = currentRead
		}
	}

	for lastRead <= ringbuffer.headPointer {
		runtime.Gosched()
		lastRead = ringbuffer.headPointer + ringbuffer.length
		for i = 0; i < ringbuffer.readerCount; i++ {

			if ringbuffer.readersPosition[i] == nil {
				continue
			}
			currentRead = *ringbuffer.readersPosition[i] + ringbuffer.length
			if currentRead < lastRead {
				lastRead = currentRead
			}
		}
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

func (consumer *Consumer[T]) Get() T {
	return consumer.ring.readIndex(consumer.id)
}
