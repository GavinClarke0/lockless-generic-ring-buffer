package lockless_generic_ring_buffer

import (
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	BufferSizeStandard = 100
	BufferSizeSmall    = 10
	BufferSizeTiny     = 2
)

func TestGetsAreSequentiallyOrdered(t *testing.T) {

	//ring := make([]int, 10, 10)

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	consumer, _ := buffer.CreateConsumer()

	for _, value := range messages {
		buffer.Write(value)
	}

	for i, _ := range messages {
		value := consumer.Get()

		if value != messages[i] {
			t.FailNow()
		}
	}
}

// test adding a consumer mid work
func TestNewConsumerReadsFromCurrentWritePosition(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	consumer1, _ := buffer.CreateConsumer()

	for _, value := range messages[:5] {
		buffer.Write(value)
	}

	consumer2, _ := buffer.CreateConsumer()

	for _, value := range messages[5:] {
		buffer.Write(value)
	}

	for _, value := range messages {

		getValue := consumer1.Get()

		if getValue != value {
			t.Fail()
		}
	}

	// test it reads froms
	for _, value := range messages[5:] {
		getValue := consumer2.Get()

		if getValue != value {
			t.Fail()
		}
	}

}

// test adding a consumer mid work
func TestRemovingConsumerDoesNotBlockNewWrites(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	consumer1, _ := buffer.CreateConsumer()
	consumer2, _ := buffer.CreateConsumer()

	for _, value := range messages[:5] {
		buffer.Write(value)
	}

	consumer2.Remove()

	for _, value := range messages[5:] {
		buffer.Write(value)
	}

	for _, value := range messages {

		getValue := consumer1.Get()

		if getValue != value {
			t.Fail()
		}
	}

	if buffer.readerActiveFlags[1] != 0 {
		t.Fail()
	}
}

// Test order is still preserved with simultaneous reading writing
func TestConcurrentGetsAreSequentiallyOrdered(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 100000; i++ {
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
				t.Fail()
			}
			i = j
		}
	}()
	wg.Wait()
}

// Test order is still preserved with simultaneous reading writing
func TestConcurrentGetsAreSequentiallyOrderedMinibuffer(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeTiny, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 100000; i++ {
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
				t.Fail()
			}
			i = j
		}
	}()
	wg.Wait()
}

func TestConcurrentStringsGetsAreSequentiallyOrderedWithMultiConsumer(t *testing.T) {

	var buffer = CreateBuffer[string](BufferSizeSmall, 10)

	var wg sync.WaitGroup
	messages := []string{}

	for i := 0; i < 100000; i++ {
		token := make([]byte, 16)
		rand.Read(token)
		messages = append(messages, string(token))
	}

	consumer1, _ := buffer.CreateConsumer()
	consumer2, _ := buffer.CreateConsumer()
	consumer3, _ := buffer.CreateConsumer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer.Write(value)
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer1.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer2.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer3.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()
	wg.Wait()
}

// Test all values are read in order
func TestConcurrentGetsAreSequentiallyOrderedWithMultiConsumer(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 1000000; i++ {
		messages = append(messages, i)
	}

	consumer1, _ := buffer.CreateConsumer()
	consumer2, _ := buffer.CreateConsumer()
	consumer3, _ := buffer.CreateConsumer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer.Write(value)
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer1.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer2.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()

	wg.Add(1)
	go func() {
		//i := -1
		defer wg.Done()
		for _, value := range messages {
			j := consumer3.Get()
			if j != value {
				t.Fail()
			}
			//i = j
		}
	}()
	wg.Wait()
}

// Test all values are read in order
func TestConcurrentAddRemoveConsumerDoesNotBlockWrites(t *testing.T) {

	var buffer = CreateBuffer[int](BufferSizeStandard, 10)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 10000; i++ {
		messages = append(messages, i)
	}

	consumer1, _ := buffer.CreateConsumer()

	failIfDeadLock(t)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			buffer.Write(value)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, value := range messages {
			j := consumer1.Get()
			if j != value {
				fmt.Println("bad value 1")
				t.Fail()
			}
		}
	}()

	wg.Add(1)
	go func() {
		consumer2, _ := buffer.CreateConsumer()

		defer wg.Done()
		for _, _ = range messages[:500] {
			consumer2.Get()
		}

		consumer2.Remove()
	}()

	wg.Wait()
}

func failIfDeadLock(t *testing.T) {
	// fail if routine is blocking
	go time.AfterFunc(1*time.Second, func() {
		fmt.Println("DeadLock")
		t.FailNow()
	})
}
