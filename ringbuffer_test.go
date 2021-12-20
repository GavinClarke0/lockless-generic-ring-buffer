package ringbuffer

import (
	"fmt"
	"sync"
	"testing"
)

func TestIntMinTableDriven(t *testing.T) {

	//ring := make([]int, 10, 10)

	var buffer = CreateBuffer[int](10, false)

	messages := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	consumer := buffer.CreateConsumer()

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
func TestIntMin(t *testing.T) {

	//ring := make([]int, 10, 10)

	var buffer = CreateBuffer[int](10, false)

	var wg sync.WaitGroup
	messages := []int{}

	for i := 0; i < 10000; i++ {
		messages = append(messages, i)
	}

	consumer := buffer.CreateConsumer()

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
				t.Fail()
			}
			i = j
		}
	}()

	wg.Wait()

	x := 0
	fmt.Println(x)
}
