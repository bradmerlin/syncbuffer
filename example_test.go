package syncbuffer_test

import (
	"log"
	"sync"
	"time"

	"github.com/ny0m/syncbuffer"
)

func Example() {
	// New buffer with space for 2 items that adds messages every millisecond.
	sb := syncbuffer.NewSyncBuffer(time.Second/2, 2)

	var wg sync.WaitGroup
	wg.Add(1)

	// Start streaming somewhere.
	go func() {
		defer wg.Done()

		s := syncbuffer.NewStreamer(sb)
		for p := range s.Stream() {
			log.Println(p)
		}
	}()

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
