package eventsource

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"
)

// RunDatanodeEventSource is the main function of `eventsource` package
func RunDatanodeEventSource(eventsFile string, port uint, closeConnection bool, timeBetweenBlocks time.Duration,
	consoleLogFormat string,
) error {
	eventSource, err := newDatanodeEventSource(eventsFile, timeBetweenBlocks, port, consoleLogFormat)
	if err != nil {
		return fmt.Errorf("failed to create event source: %w", err)
	}

	err = eventSource.socketClient.connect()
	defer eventSource.socketClient.close()
	if err != nil {
		return fmt.Errorf("failed to connect socket client source: %w", err)
	}

	err = eventSource.sendEvents()
	if err != nil {
		return fmt.Errorf("failed to send events: %w", err)
	}

	if !closeConnection {
		for {
			time.Sleep(time.Second)
		}
	}

	return nil
}

type dataNodeEventSource struct {
	socketClient      *socketClient
	eventsFile        string
	timeBetweenBlocks time.Duration
	logEventToConsole func(e *eventspb.BusEvent)
}

func newDatanodeEventSource(eventsFile string, timeBetweenBlocks time.Duration,
	port uint, logFormat string) (*dataNodeEventSource, error,
) {
	filePath, err := filepath.Abs(eventsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to determine absolute path of file %s: %w", eventsFile, err)
	}
	fmt.Printf("creating event source for events at: %s\n", filePath)

	address := fmt.Sprintf("tcp://0.0.0.0:%d", port)
	sc, err := newSocketClient(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket client for address %s: %w", address, err)
	}

	logEventToConsole, err := stream.NewLogEventToConsoleFn(logFormat)
	if err != nil {
		return nil, err
	}

	return &dataNodeEventSource{
		socketClient:      sc,
		eventsFile:        eventsFile,
		timeBetweenBlocks: timeBetweenBlocks,
		logEventToConsole: logEventToConsole,
	}, nil
}

func (e dataNodeEventSource) sendEvents() error {
	fi, err := os.Open(e.eventsFile)
	defer fi.Close()

	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", e.eventsFile, err)
	}

	return sendAllEvents(e.socketClient.send, fi, 0, e.timeBetweenBlocks, e.logEventToConsole)
}
