package eventrate

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"

	"github.com/gdamore/tcell/v2"
)

type counter struct {
	eventType          eventspb.BusEventType
	totalEvents        uint64
	totalSecond        uint64
	minEventsPerSecond uint64
	maxEventsPerSecond uint64
	avgEventsPerSecond uint64
}

var (
	server              string
	ts                  tcell.Screen
	greyStyle           tcell.Style
	whiteStyle          tcell.Style
	currentEventCounts  map[eventspb.BusEventType]uint64   = map[eventspb.BusEventType]uint64{}
	historicEventCounts map[eventspb.BusEventType]*counter = map[eventspb.BusEventType]*counter{}
	totals              counter
	currentEvents       uint64
	mu                  sync.Mutex
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

// UpdateCounters updates the current event values
func UpdateCounters() {
	// Go through every current event counter and apply them to the historic counters
	for key, value := range currentEventCounts {
		// Update the historic per second values
		historic := historicEventCounts[key]
		historic.totalEvents += value
		historic.totalSecond++
		historic.minEventsPerSecond = min(historic.minEventsPerSecond, value)
		historic.maxEventsPerSecond = max(historic.maxEventsPerSecond, value)
		historic.avgEventsPerSecond = historic.totalEvents / historic.totalSecond

		// Clear the current second counter
		currentEventCounts[key] = 0
	}

	// Same for the totals counter
	totals.totalEvents += currentEvents
	totals.totalSecond++
	totals.minEventsPerSecond = min(totals.minEventsPerSecond, currentEvents)
	totals.maxEventsPerSecond = max(totals.maxEventsPerSecond, currentEvents)
	totals.avgEventsPerSecond = totals.totalEvents / totals.totalSecond
	currentEvents = 0
}

// Run is the main function of `eventpersister` package
func Run(serverAddr string) error {
	server = serverAddr
	flag.Parse()

	if len(serverAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	handleEvent := func(e *eventspb.BusEvent) {
		mu.Lock()
		if count, ok := currentEventCounts[e.Type]; ok {
			currentEventCounts[e.Type] = count + 1
		} else {
			currentEventCounts[e.Type] = 1
			// Create the historic counter at the same time
			historicEventCounts[e.Type] = &counter{eventType: e.Type}
		}
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := stream.ReadEvents(ctx, cancel, &wg, 0, "", "", serverAddr, handleEvent, true, nil); err != nil {
		return fmt.Errorf("error reading events: %v", err)
	}

	initialiseScreen()

	for i := 0; i < 100; i++ {
		// Wait a second for the next update
		time.Sleep(time.Second)

		// Recalculate the counters
		mu.Lock()
		UpdateCounters()

		// Redraw the screen
		drawScreen()
		mu.Unlock()
	}

	stream.WaitSig(ctx, cancel)
	wg.Wait()

	return nil
}
