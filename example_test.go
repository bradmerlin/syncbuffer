package syncbuffer_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/bradmerlin/syncbuffer"
)

func Example() {
	// New buffer with space for 2 items that adds messages every millisecond.
	sb := syncbuffer.NewSyncBuffer(time.Millisecond, 2)

	var wg sync.WaitGroup
	wg.Add(1)

	// Start streaming somewhere.
	go func() {
		defer wg.Done()

		s := syncbuffer.NewStreamer(sb)
		for p := range s.Stream() {
			fmt.Println(p)
		}
	}()

	time.Sleep(time.Millisecond)

	for i := 0; i < 10; i++ {
		// Add will block until a millisecond has passed.
		sb.Add([]byte{byte(i)})
	}

	// Close the buffer and stop its streamers.
	sb.Close()

	// Ensure that the program waits for the streamer to finish.
	wg.Wait()

	// Output:
	// [0]
	// [1]
	// [2]
	// [3]
	// [4]
	// [5]
	// [6]
	// [7]
	// [8]
	// [9]
}
