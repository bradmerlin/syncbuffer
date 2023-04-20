# syncbuffer

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/ny0m/syncbuffer/)

Syncbuffer's structs are useful for applications where multiple separate readers need to 
be kept roughly synchronised with a source; for example, audio or video streaming.

## Example
```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/ny0m/syncbuffer"
)

func main() {
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
```
