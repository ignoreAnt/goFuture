package main

import (
	"fmt"
	"sync"
	"time"
)

func main()  {
	testWaitOnGet()
	testTimeout()
}

// Test function to verify that test is received properly
func testWaitOnGet()  {
	receiver := make(chan interface{})
	future := createFuture(receiver, time.Duration(5*time.Second))
	var result interface{}
	var err error
	var waitG sync.WaitGroup
	waitG.Add(1)

	go func() {
		defer waitG.Done()
		result, err = future.get()
	}()

	receiver <- "test"
	waitG.Wait()
	fmt.Println("Error is : ", err)
	fmt.Println("Result is : ", result)
}

// Test function to verify the timeout
func testTimeout()  {
	receiver := make(chan interface{})
	future := createFuture(receiver, time.Duration(0))
	var result interface{}
	var err error
	result, err = future.get()
	fmt.Println("Error is : ", err)
	fmt.Println("Result is : ", result)
}

// Receiver is a channel where future will get the results
type Receiver <-chan interface{}

// Future is a struct that can be used to perform async tasks
type Future struct {
	isInitiated bool
	object      interface{}
	err         error
	lock        sync.Mutex
	waitG       sync.WaitGroup
}

// Get will either get the results if it exists or will wait for the result till its ready
func (future *Future) get() (interface{}, error) {
	future.lock.Lock()

	if future.isInitiated {
		future.lock.Unlock()
		return future.object, future.err
	}

	future.lock.Unlock()

	future.waitG.Wait()
	return future.object, future.err
}

// Sets the various attributes to the future struct
func (future *Future) set(obj interface{}, err error) {
	future.lock.Lock()
	future.isInitiated = true
	future.object = obj
	future.err = err
	future.lock.Unlock()
	future.waitG.Done()
}

// Listener which listens to the result and
func listener(future *Future, rc Receiver, timeout time.Duration, group *sync.WaitGroup) {
	group.Done()
	select {
	case obj := <-rc:
		future.set(obj, nil)
	case <-time.After(timeout):
		future.set(nil, fmt.Errorf("timed out after %f seconds", timeout.Seconds()))
	}
}

// Creates a new instance of future with a channel, duration
func createFuture(receiver Receiver, duration time.Duration) *Future {
	future := &Future{}
	future.waitG.Add(1)

	var waitG sync.WaitGroup
	waitG.Add(1)

	go listener(future, receiver, duration, &waitG)
	waitG.Wait()

	return future
}
