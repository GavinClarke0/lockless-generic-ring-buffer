package ringbuffer

import (
	"runtime"
	"sync/atomic"
)

var (
	MaxUint32    = ^uint32(0)
	MaxConsumers = 100
)

type RingBuffer[T any] struct {
	buffer          []T
	length          uint32
	headPointer     uint32 // next position to write
	readerCount     uint32
	readersPosition [100]uint32
}

type Consumer[T any] struct {
	ring *RingBuffer[T]
	id   uint32
}

func CreateBuffer[T any](size uint32) RingBuffer[T] {

	var buffer = [100]uint32{}

	for i, _ := range buffer {
		buffer[i] = MaxUint32
	}

	return RingBuffer[T]{
		buffer:          make([]T, size, size),
		length:          size,
		headPointer:     0,
		readersPosition: buffer,
	}
}

func (ringbuffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	var position = atomic.AddUint32(&ringbuffer.readerCount, 1)
	ringbuffer.readersPosition[position-1] = ringbuffer.headPointer - 1

	return Consumer[T]{
		id:   position - 1,
		ring: ringbuffer,
	}, nil
}

func (ringbuffer *RingBuffer[T]) Write(value T) {

	var lastRead uint32
	var i uint32

	/*
		Non-critical path, we are blocking until the all at least one space is available in the buffer
	*/
	lastRead = ringbuffer.readersPosition[0]
	for i = 1; i <= ringbuffer.readerCount; i++ {
		if ringbuffer.readersPosition[i] < lastRead {
			lastRead = ringbuffer.readersPosition[i]
		}
	}

	for lastRead+ringbuffer.length <= ringbuffer.headPointer {
		runtime.Gosched()
		lastRead = ringbuffer.readersPosition[0]
		for i = 1; i < ringbuffer.readerCount; i++ {
			if ringbuffer.readersPosition[i] < lastRead {
				lastRead = ringbuffer.readersPosition[i]
			}
		}
	}

	ringbuffer.buffer[(ringbuffer.headPointer)%ringbuffer.length] = value
	atomic.AddUint32(&ringbuffer.headPointer, 1)
}

func (ringbuffer *RingBuffer[T]) readIndex(consumerId uint32) T {

	var newIndex = atomic.AddUint32(&ringbuffer.readersPosition[consumerId], 1)

	// yield until work is available
	for newIndex >= ringbuffer.headPointer {
		runtime.Gosched()
	}
	return ringbuffer.buffer[newIndex%ringbuffer.length]
}

func (consumer *Consumer[T]) Get() T {
	return consumer.ring.readIndex(consumer.id)
}
