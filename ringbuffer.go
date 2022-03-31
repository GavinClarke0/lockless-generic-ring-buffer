package locklessgenericringbuffer

import (
	"errors"
	"runtime"
	"sync/atomic"
)

var (
	MaxConsumerError  = errors.New("max amount of consumers reached cannot create any more")
	InvalidBufferSize = errors.New("buffer must be of size 2^n")
)

type RingBuffer[T any] struct {
	length            uint32
	bitWiseLength     uint32
	headPointer       uint32 // next position to write
	maximumConsumerId uint32
	maxConsumers      int
	buffer            []T
	readerPointers    []uint32
	readerActiveFlags []uint32
}

type Consumer[T any] struct {
	ring *RingBuffer[T]
	id   uint32
}

func CreateBuffer[T any](size uint32, maxConsumers uint32) (RingBuffer[T], error) {

	if size&(size-1) != 0 {
		return RingBuffer[T]{}, InvalidBufferSize
	}

	return RingBuffer[T]{
		buffer:            make([]T, size+1, size+1),
		length:            size,
		bitWiseLength:     size - 1,
		headPointer:       0,
		maximumConsumerId: 0,
		maxConsumers:      int(maxConsumers),
		readerPointers:    make([]uint32, maxConsumers),
		readerActiveFlags: make([]uint32, maxConsumers),
	}, nil
}

/*
CreateConsumer

Create a consumer by assigning it the id of the first empty position in the consumerPosition array. A nil value represents
an unclaimed/not used consumer.

*/
func (ringbuffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	for newConsumerId, _ := range ringbuffer.readerActiveFlags {
		if atomic.CompareAndSwapUint32(&ringbuffer.readerActiveFlags[newConsumerId], 0, 1) {

			if uint32(newConsumerId) >= ringbuffer.maximumConsumerId {
				atomic.AddUint32(&ringbuffer.maximumConsumerId, 1)
			}

			ringbuffer.readerPointers[newConsumerId] = atomic.LoadUint32(&ringbuffer.headPointer) - 1
			atomic.StoreUint32(&ringbuffer.readerActiveFlags[newConsumerId], 1)

			return Consumer[T]{
				id:   uint32(newConsumerId),
				ring: ringbuffer,
			}, nil
		}
	}

	return Consumer[T]{}, MaxConsumerError
}

func (ringbuffer *RingBuffer[T]) removeConsumer(consumerId uint32) {

	atomic.StoreUint32(&ringbuffer.readerActiveFlags[consumerId], 0)
	atomic.CompareAndSwapUint32(&ringbuffer.maximumConsumerId, consumerId, ringbuffer.maximumConsumerId-1)
}

func (consumer *Consumer[T]) Remove() {
	consumer.ring.removeConsumer(consumer.id)
}

func (consumer *Consumer[T]) Get() T {
	return consumer.ring.readIndex(consumer.id)
}

func (ringbuffer *RingBuffer[T]) Write(value T) {

	var lastTailReaderPointerPosition uint32
	var currentReadPosition uint32
	var i uint32
	/*
		We are blocking until the all at least one space is available in the buffer to write.

		As overflow properties of uint32 are utilized to ensure slice index boundaries are adhered too we add the length
		of buffer to current consumer read positions allowing us to determine the least read consumer.

		For example: buffer of size 2

		uint8 head = 1
		uint8 tail = 255
		tail + 2 => 1 with overflow, same as buffer
	*/
	for {
		lastTailReaderPointerPosition = atomic.LoadUint32(&ringbuffer.headPointer) + ringbuffer.length

		for i = 0; i < atomic.LoadUint32(&ringbuffer.maximumConsumerId); i++ {

			if atomic.LoadUint32(&ringbuffer.readerActiveFlags[i]) == 1 {
				currentReadPosition = atomic.LoadUint32(&ringbuffer.readerPointers[i]) + ringbuffer.length

				if currentReadPosition < lastTailReaderPointerPosition {
					lastTailReaderPointerPosition = currentReadPosition
				}
			}
		}

		if lastTailReaderPointerPosition > ringbuffer.headPointer {

			ringbuffer.buffer[ringbuffer.headPointer&ringbuffer.bitWiseLength] = value
			atomic.AddUint32(&ringbuffer.headPointer, 1)
			return
		}
		runtime.Gosched()
	}
}

func (ringbuffer *RingBuffer[T]) readIndex(consumerId uint32) T {

	var newIndex = atomic.AddUint32(&ringbuffer.readerPointers[consumerId], 1)

	// yield until work is available
	for newIndex >= atomic.LoadUint32(&ringbuffer.headPointer) {
		runtime.Gosched()
	}
	return ringbuffer.buffer[newIndex&ringbuffer.bitWiseLength]
}
