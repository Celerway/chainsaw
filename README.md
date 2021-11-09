# Chainsaw

Circular buffer logging framework for Go. 

This logging library will keep the last N log messages around. This allows you to 
dump the logs at trace or debug level if you encounter an error, contextualizing
the error.

It also allows you to start a stream of log messages, making it suitable for 
applications which have sub-applications where you might wanna look at different 
streams of logs. Say in an application which has a bunch of goroutines which have
more or less independent log.

## Example usage

```go
package main

import (
	"context"
	"fmt"
	log "github.com/celerway/chainsaw"
	"os"
	"time"
)

func main() {
	log.Info("Application starting up")
	log.Trace("This is a trace message. Doesn't get printed directly")
	logMessages := log.GetMessages(log.TraceLevel)
	for i, mess := range logMessages {
		fmt.Println("Fetched message: ", i, ":", mess.Content)
	}
	log.RemoveWriter(os.Stdout) // stop writing to stdout
	time.Sleep(10 * time.Millisecond)
	log.Info("Doesn't show up on screen.")
	ctx, cancel := context.WithCancel(context.Background())
	stream := log.GetStream(ctx)
	go func() {
		for mess := range stream {
			fmt.Println("From stream: ", mess.Content)
		}
		fmt.Println("Log stream is closed")
	}()
	log.Info("Should reach the stream above")
	cancel()
	time.Sleep(time.Second)
}
```
