package ringbuffer

import (
	"sync"
	"testing"
)

/*
General Benchmark to compare reading from channels vrs the ring buffer.

Note there is heavy over head for syncing the routines in both and is not accurate beyond a general comparison.
*/
func BenchmarkConsumerConcurrentReadWrite(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ConsumerConcurrentReadWrite(100000)
	}
}

func BenchmarkChannelsConcurrentReadWrite(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ChannelsConcurrentReadWrite(100000)
	}
}

func ConsumerConcurrentReadWrite(n int) {

	var buffer = CreateBuffer[int](BufferSize, 10)

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

	var buffer = make(chan int, BufferSize)

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
