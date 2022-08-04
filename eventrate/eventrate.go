package eventrate

import (
	"context"
	"fmt"
	"sync"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"
)

func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// Run is the main function of `eventpersister` package
func Run(serverAddr string) error {
	var	eventsThisSecond    uint64
	var historicEvents      []uint64
	var mu                  sync.Mutex

	if len(serverAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	handleEvent := func(e *eventspb.BusEvent) {
		mu.Lock()
		eventsThisSecond++
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := stream.ReadEvents(ctx, cancel, &wg, 0, "", "", serverAddr, handleEvent, true, nil); err != nil {
		return fmt.Errorf("error reading events: %v", err)
	}

	for {
		time.Sleep(time.Second)
		mu.Lock()
		historicEvents = append(historicEvents, eventsThisSecond)
		eventsThisSecond = 0
		mu.Unlock()

		// If we have more than 10 historic counts, remove the last
		if len(historicEvents) > 10 {
			historicEvents = historicEvents[1:]
		}

		// Calculate values
		minimum := historicEvents[0]
		maximum := historicEvents[0]
		var total uint64
		for _, i := range historicEvents {
			minimum = min(minimum,i)
			maximum = max(maximum,i)
			total += i
		}
		average := total/uint64(len(historicEvents))
		fmt.Printf("Events per second: (")
		for i:=len(historicEvents)-1;i>0;i-- {
			fmt.Printf("%d, ", historicEvents[i])
		}
		fmt.Printf("%d) Min:%d Max:%d Avg:%d            \r", historicEvents[0], minimum, maximum, average)
	}
	return nil
}
