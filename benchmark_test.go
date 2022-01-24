package lockless_generic_ring_buffer

import (
	"sync"
	"testing"
)

const (
	MessageCountLarge  = 10000000
	MessageCountMedium = 100000
	MessageCountSmall  = 1000
)

func BenchmarkConsumerSequentialReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountLarge)
	}
}

func BenchmarkChannelsSequentialReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountLarge)
	}
}

func BenchmarkConsumerSequentialReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountMedium)
	}
}

func BenchmarkChannelsSequentialReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountMedium)
	}
}

func BenchmarkConsumerSequentialReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountSmall)
	}
}

func BenchmarkChannelsSequentialReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountSmall)
	}
}

func ConsumerSequentialReadWrite(n int) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)
	consumer, _ := buffer.CreateConsumer()

	for i := 0; i < n; i++ {
		buffer.Write(i)
		consumer.Get()
	}
}

func ChannelsSequentialReadWrite(n int) {

	var buffer = make(chan int, BufferSizeStandard)

	for i := 0; i < n; i++ {
		buffer <- i
		<-buffer
	}
}

/*
General Benchmark to compare concurrent reading from channels vrs the ring buffer.
Note there is heavy over head for syncing the routines in both and is not accurate beyond a general comparison.
*/
func BenchmarkConsumerConcurrentReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(MessageCountLarge)
	}
}

func BenchmarkChannelsConcurrentReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountLarge)
	}
}

func BenchmarkConsumerConcurrentReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(MessageCountMedium)
	}
}

func BenchmarkChannelsConcurrentReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountMedium)
	}
}
func BenchmarkConsumerConcurrentReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(MessageCountSmall)
	}
}

func BenchmarkChannelsConcurrentReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountSmall)
	}
}

func ConsumerConcurrentReadWrite(n int) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < n; i++ {
		messages = append(messages, i)
	}

	consumer, _ := buffer.CreateConsumer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer.Write(value)
		}
	}()

	i := -1

	wg.Add(1)
	go func() {

		defer wg.Done()
		for _, _ = range messages {
			j := consumer.Get()
			if j != i+1 {
				panic("data is inconsistent")
			}
			i = j
		}
	}()
	wg.Wait()
}

func ChannelsConcurrentReadWrite(n int) {

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < n; i++ {
		messages = append(messages, i)
	}

	var buffer = make(chan int, BufferSizeStandard)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer <- value
		}
	}()

	i := -1

	wg.Add(1)
	go func() {

		defer wg.Done()
		for _, _ = range messages {
			j := <-buffer
			if j != i+1 {
				panic("data is inconsistent")
			}
			i = j
		}
	}()
	wg.Wait()
}
