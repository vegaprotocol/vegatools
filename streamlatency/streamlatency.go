package streamlatency

import (
	"context"
	"fmt"
	"sync"
	"time"

	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"
)

// Opts are the command line options passed to the sub command
type Opts struct {
	ServerAddr1 string
	ServerAddr2 string
	ReportMode  bool
}

const colorRed = "\033[0;31m"
const colorGreen = "\033[0;32m"

// Run is the main function of `eventrate` package
func Run(opts Opts) error {
	var (
		// Allow up to 100 blocks of historic timings
		serverTimings1 []int64 = make([]int64, 100)
		serverTimings2 []int64 = make([]int64, 100)
		lastHeight1    uint64
		lastHeight2    uint64
		mu             sync.Mutex
	)

	handleEvent1 := func(e *eventspb.BusEvent) {
		mu.Lock()
		switch e.Type {
		case eventspb.BusEventType_BUS_EVENT_TYPE_END_BLOCK:
			eb := e.GetEndBlock()
			lastHeight1 = eb.Height
			now := time.Now().UnixMilli()
			offset := lastHeight1 % 100
			// If we already have a non zero value in the other list we can compare
			if serverTimings2[offset] != 0 {
				fmt.Printf("%s%s is behind %s by %d milliseconds    \r", colorRed, opts.ServerAddr1, opts.ServerAddr2, now-serverTimings2[offset])
				if opts.ReportMode {
					fmt.Println()
				}
				serverTimings2[offset] = 0
			} else {
				serverTimings1[offset] = now
			}
		}
		mu.Unlock()
	}

	handleEvent2 := func(e *eventspb.BusEvent) {
		mu.Lock()
		switch e.Type {
		case eventspb.BusEventType_BUS_EVENT_TYPE_END_BLOCK:
			eb := e.GetEndBlock()
			lastHeight2 = eb.Height
			now := time.Now().UnixMilli()
			offset := lastHeight2 % 100
			// If we already have a non zero value in the other list we can compare
			if serverTimings1[offset] != 0 {
				fmt.Printf("%s%s is behind %s by %d milliseconds    \r", colorGreen, opts.ServerAddr2, opts.ServerAddr1, now-serverTimings1[offset])
				if opts.ReportMode {
					fmt.Println()
				}
				serverTimings1[offset] = 0
			} else {
				serverTimings2[offset] = now
			}
		}
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	types := []string{"BUS_EVENT_TYPE_END_BLOCK"}

	// Connect to the 2 event streams and start processing the incoming events
	if err := stream.ReadEvents(ctx, cancel, &wg, 0, "", "", opts.ServerAddr1, handleEvent1, true, types); err != nil {
		return fmt.Errorf("error reading events from stream 1: %v", err)
	}

	if err := stream.ReadEvents(ctx, cancel, &wg, 0, "", "", opts.ServerAddr2, handleEvent2, true, types); err != nil {
		return fmt.Errorf("error reading events from stream 2: %v", err)
	}

	for {
	}
}
