package eventrate

import (
	"context"
	"fmt"
	"sync"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"
	"google.golang.org/protobuf/proto"
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

type Data struct {
	Events uint64
	Bytes  uint64
}

func fixUnits(bytes uint64) string {
	if bytes > 1000000 {
		return fmt.Sprintf("%dMB", bytes/(1024*1024))
	} else if bytes > 1000 {
		return fmt.Sprintf("%dKB", bytes/1024)
	}
	return fmt.Sprintf("%dB", bytes)
}

// Run is the main function of `eventrate` package
func Run(serverAddr string) error {
	var dataThisSecond Data
	var historicData []Data
	var mu sync.Mutex

	if len(serverAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	handleEvent := func(e *eventspb.BusEvent) {
		mu.Lock()
		dataThisSecond.Events++
		dataThisSecond.Bytes += uint64(proto.Size(e))
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
		historicData = append(historicData, dataThisSecond)
		dataThisSecond.Events = 0
		dataThisSecond.Bytes = 0
		mu.Unlock()

		// If we have more than 10 historic counts, remove the last
		if len(historicData) > 10 {
			historicData = historicData[1:]
		}

		// Calculate values
		minEvents := historicData[0].Events
		maxEvents := historicData[0].Events
		minBytes := historicData[0].Bytes
		maxBytes := historicData[0].Bytes
		var totalEvents uint64
		var totalBytes uint64
		for _, i := range historicData {
			minEvents = min(minEvents, i.Events)
			maxEvents = max(maxEvents, i.Events)
			totalEvents += i.Events
			minBytes = min(minBytes, i.Bytes)
			maxBytes = max(maxBytes, i.Bytes)
			totalBytes += i.Bytes
		}
		avgEvents := totalEvents / uint64(len(historicData))
		avgBytes := totalBytes / uint64(len(historicData))
		fmt.Printf("Events:Bandwidth (")
		for i := len(historicData) - 1; i > 0; i-- {
			fmt.Printf("[%d:%s], ", historicData[i].Events, fixUnits(historicData[i].Bytes))
		}
		fmt.Printf("[%d:%s]) Min:[%d:%s] Max:[%d:%s] Avg:[%d:%s]            \r",
			historicData[0].Events, fixUnits(historicData[0].Bytes),
			minEvents, fixUnits(minBytes), maxEvents, fixUnits(maxBytes), avgEvents, fixUnits(avgBytes))
	}
}
