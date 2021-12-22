package ringbuffer

import (
	"fmt"
	"sync"
	"testing"
)

func TestSequentialInt(t *testing.T) {

	//ring := make([]int, 10, 10)

	var buffer = CreateBuffer[int](10)

	messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	consumer, _ := buffer.CreateConsumer()

	for _, value := range messages {
		buffer.Write(value)
	}

	for _, _ = range messages {
		x := consumer.Get()
		fmt.Println(x)
	}
}

/*
Test order is still preserved with simultaneous reading writing
*/
func TestConcurrentSingleProducerConsumer(t *testing.T) {

	var buffer = CreateBuffer[int](100)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 1000000; i++ {
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
			if j < i {
				fmt.Println("access order invalid")
				t.Fail()
			}
			i = j
		}
	}()
	wg.Wait()
}

func TestConcurrentSingleProducerMultiConsumer(t *testing.T) {

	var buffer = CreateBuffer[int](50)

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

		i := 0

		defer wg.Done()
		for _, _ = range messages {
			j := consumer1.Get()

			//fmt.Println(j)
			if j < i {
				//fmt.Println("fail")
				t.Fail()
			}
			i = j
		}
	}()

	wg.Add(1)
	go func() {

		i := 0
		defer wg.Done()
		for _, _ = range messages {
			j := consumer2.Get()
			if j < i {
				fmt.Println("fail")
				t.Fail()
			}
			i = j
		}
	}()

	wg.Add(1)
	go func() {

		i := 0
		defer wg.Done()
		for _, _ = range messages {
			j := consumer3.Get()
			if j < i {
				fmt.Println("fail")
				t.Fail()
			}
			i = j
		}
	}()
	wg.Wait()
}
