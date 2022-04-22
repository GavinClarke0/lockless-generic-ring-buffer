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
	headIndex         uint32 // next position to write
	nextReaderIndex   uint32
	maxReaders        int
	buffer            []T
	readerIndexes     []uint32
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
		buffer:            make([]T, size, size),
		length:            size,
		bitWiseLength:     size - 1,
		headIndex:         0,
		nextReaderIndex:   0,
		maxReaders:        int(maxConsumers),
		readerIndexes:     make([]uint32, maxConsumers),
		readerActiveFlags: make([]uint32, maxConsumers),
	}, nil
}

/*
CreateConsumer

Create a consumer by assigning it the id of the first empty position in the consumerPosition array. A nil value represents
an unclaimed/not used consumer.
*/
func (buffer *RingBuffer[T]) CreateConsumer() (Consumer[T], error) {

	for readerIndex, _ := range buffer.readerActiveFlags {
		if atomic.CompareAndSwapUint32(&buffer.readerActiveFlags[readerIndex], 0, 2) {

			atomic.CompareAndSwapUint32(&buffer.nextReaderIndex, uint32(readerIndex), uint32(readerIndex)+1)

			// as read state is set to 2, we can afford to non atomically set readIndex
			buffer.readerIndexes[readerIndex] = atomic.LoadUint32(&buffer.headIndex)
			atomic.StoreUint32(&buffer.readerActiveFlags[readerIndex], 1)

			return Consumer[T]{
				id:   uint32(readerIndex),
				ring: buffer,
			}, nil
		}
	}

	return Consumer[T]{}, MaxConsumerError
}

func (buffer *RingBuffer[T]) removeConsumer(readerId uint32) {
	atomic.StoreUint32(&buffer.readerActiveFlags[readerId], 0)
	atomic.CompareAndSwapUint32(&buffer.nextReaderIndex, readerId-1, buffer.nextReaderIndex-1)
}

func (consumer *Consumer[T]) Remove() {
	consumer.ring.removeConsumer(consumer.id)
}

func (consumer *Consumer[T]) Get() T {
	return consumer.ring.readIndex(consumer.id)
}

func (buffer *RingBuffer[T]) Write(value T) {

	var delta uint32
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

		nextReaderIndex := atomic.LoadUint32(&buffer.nextReaderIndex)
		write := true

		for i = 0; i < nextReaderIndex; i++ {
			if atomic.LoadUint32(&buffer.readerActiveFlags[i]) == 1 {
				delta = atomic.LoadUint32(&buffer.readerIndexes[i]) + buffer.length

				// only true if the delta between at least one reader and the writer is equal to the size of the buffer
				if delta == buffer.headIndex {
					write = false
				}
			}
		}

		if write {
			buffer.buffer[(buffer.headIndex+1)&buffer.bitWiseLength] = value
			atomic.AddUint32(&buffer.headIndex, 1)
			return
		}
		runtime.Gosched()
	}
}

func (buffer *RingBuffer[T]) readIndex(readerIndex uint32) T {

	var newIndex = buffer.readerIndexes[readerIndex] + 1
	// yield until work is available
	for newIndex > atomic.LoadUint32(&buffer.headIndex) {
		runtime.Gosched()
	}

	value := buffer.buffer[newIndex&buffer.bitWiseLength]
	atomic.AddUint32(&buffer.readerIndexes[readerIndex], 1)
	return value
}
