package ringbuffer

import (
	"errors"
	"sync"
)

var (
	ErrorIsFull      = errors.New("ring buffer is full")
	ErrorIsEmpty     = errors.New("ring buffer is empty")
	ErrorAcquireLock = errors.New("no lock to accquire")
)

type RingBuffer[T any] struct {
	buffer       []T
	lock         sync.RWMutex
	length       int
	writePointer int // next position to write
	tailPointer  int
	overwrite    bool
}

type Consumer[T any] struct {
	readPointer int
	lock        sync.RWMutex
	ring        *RingBuffer[T]
	size        int
}

func CreateBuffer[T any](size int, overwrite bool) RingBuffer[T] {
	return RingBuffer[T]{
		buffer:       make([]T, size, size),
		lock:         sync.RWMutex{},
		size:         size,
		writePointer: 0,
		overwrite:    overwrite,
	}
}

func (ringbuffer *RingBuffer[T]) CreateConsumer() Consumer[T] {
	return Consumer[T]{
		readPointer: ringbuffer.writePointer,
		ring:        ringbuffer,
		lock:        sync.RWMutex{},
		size:        ringbuffer.length,
	}
}

func (ringbuffer *RingBuffer[T]) Write(x T) {

	ringbuffer.lock.Lock()
	defer ringbuffer.lock.Unlock()

	var index = ringbuffer.writePointer
	ringbuffer.buffer[index] = x

	ringbuffer.writePointer = (ringbuffer.writePointer + 1) % ringbuffer.length
}

func (ringbuffer *RingBuffer[T]) readIndex(i int) T {

	ringbuffer.lock.RLock()
	defer ringbuffer.lock.RUnlock()

	return ringbuffer.buffer[i]
}

func (consumer *Consumer[T]) Get() T {
	//var value T
	var value T
	var index = consumer.readPointer

	if consumer.readPointer == consumer.ring.writePointer {
		return value
	}

	value = consumer.ring.readIndex(index)
	consumer.readPointer = (consumer.readPointer + 1) % consumer.ring.length

	return value
}

func (consumer *Consumer[T]) GetSafe() any {
	consumer.lock.Lock()
	defer consumer.lock.Unlock()

	var index = consumer.readPointer
	var value = consumer.ring.readIndex(index)

	consumer.readPointer = (consumer.readPointer + 1) % consumer.ring.length

	return value
}
