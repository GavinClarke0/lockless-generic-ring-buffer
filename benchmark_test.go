package locklessgenericringbuffer

import (
	"sync"
	"testing"
)

const (
	MessageCountLarge  = 5000000
	MessageCountMedium = 100000
	MessageCountSmall  = 1000
)

func BenchmarkConsumerSequentialReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountLarge, b)
	}
}

func BenchmarkChannelsSequentialReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountLarge, b)
	}
}

func BenchmarkConsumerSequentialReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountMedium, b)
	}
}

func BenchmarkChannelsSequentialReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountMedium, b)
	}
}

func BenchmarkConsumerSequentialReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerSequentialReadWrite(MessageCountSmall, b)
	}
}

func BenchmarkChannelsSequentialReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsSequentialReadWrite(MessageCountSmall, b)
	}
}

func ConsumerSequentialReadWrite(n int, b *testing.B) {

	b.StopTimer()
	var buffer, _ = CreateBuffer[int](BufferSizeStandard, 10)
	consumer, _ := buffer.CreateConsumer()
	b.StartTimer()

	for i := 0; i < n; i++ {
		buffer.Write(i)
		consumer.Get()
	}
}

func ChannelsSequentialReadWrite(n int, b *testing.B) {

	b.StopTimer()
	var buffer = make(chan int, BufferSizeStandard)
	b.StartTimer()

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
		ConsumerConcurrentReadWrite(MessageCountLarge, b)
	}
}

func BenchmarkChannelsConcurrentReadWriteLarge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountLarge, b)
	}
}

func BenchmarkConsumerConcurrentReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(MessageCountMedium, b)
	}
}

func BenchmarkChannelsConcurrentReadWriteMedium(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountMedium, b)
	}
}
func BenchmarkConsumerConcurrentReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(MessageCountSmall, b)
	}
}

func BenchmarkChannelsConcurrentReadWriteSmall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(MessageCountSmall, b)
	}
}

func ConsumerConcurrentReadWrite(n int, b *testing.B) {

	b.StopTimer()
	var buffer, _ = CreateBuffer[int](BufferSizeStandard, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < n; i++ {
		messages = append(messages, i)
	}

	consumer, _ := buffer.CreateConsumer()
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer.Write(value)
		}
	}()

	wg.Add(1)
	go func() {
		i := -1
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

func ChannelsConcurrentReadWrite(n int, b *testing.B) {

	b.StopTimer()
	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < n; i++ {
		messages = append(messages, i)
	}

	var buffer = make(chan int, BufferSizeStandard)
	b.StartTimer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer <- value
		}
	}()

	wg.Add(1)
	go func() {
		i := -1
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
