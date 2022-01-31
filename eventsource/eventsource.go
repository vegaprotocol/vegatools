package eventsource

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	api "code.vegaprotocol.io/protos/vega/api/v1"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

// Run is the main function of `eventsource` package
func Run(eventsFile string, port uint, closeConnection bool, timeBetweenBlocks time.Duration, consoleLogFormat string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	coreService, err := newEventSource(eventsFile, closeConnection, timeBetweenBlocks, consoleLogFormat)
	if err != nil {
		return fmt.Errorf("failed to create event source: %w", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterCoreServiceServer(grpcServer, coreService)

	fmt.Printf("listening for event subscription on port:%d\n", port)
	grpcServer.Serve(listener)

	return nil
}

type eventSource struct {
	eventsFile        string
	closeConnection   bool
	timeBetweenBlocks time.Duration
	logEventToConsole func(e *eventspb.BusEvent)
}

func newEventSource(eventsFile string, closeConnectionAfterEventsSent bool, timeBetweenBlocks time.Duration,
	logFormat string) (*eventSource, error,
) {
	filePath, err := filepath.Abs(eventsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to determine absolute path of file %s: %w", eventsFile, err)
	}
	fmt.Printf("creating event source for events in: %s\n", filePath)

	logEventToConsole, err := stream.NewLogEventToConsoleFn(logFormat)
	if err != nil {
		return nil, err
	}

	return &eventSource{
		eventsFile:        eventsFile,
		closeConnection:   closeConnectionAfterEventsSent,
		timeBetweenBlocks: timeBetweenBlocks,
		logEventToConsole: logEventToConsole,
	}, nil
}

func (e eventSource) ObserveEventBus(server api.CoreService_ObserveEventBusServer) error {
	fi, err := os.Open(e.eventsFile)
	defer fi.Close()

	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", e.eventsFile, err)
	}

	req, err := waitForObserveEventBusRequest(server)
	if err != nil {
		return fmt.Errorf("failed to wait for initial observer event request:%w", err)
	}

	sendEvents(server, fi, int(req.BatchSize), e.timeBetweenBlocks, e.logEventToConsole)

	if !e.closeConnection {
		for {
			time.Sleep(time.Second)
		}
	}

	return nil
}

func sendEvents(server api.CoreService_ObserveEventBusServer, evtFile *os.File,
	batchSize int, timeBetweenBlocks time.Duration, logEventToConsole func(e *eventspb.BusEvent),
) error {
	sizeBytes := make([]byte, 4)
	msgBytes := make([]byte, 0, 10000)
	batch := make([]*eventspb.BusEvent, 0, batchSize)
	var offset int64 = 0
	currentBlock := ""

	for {
		read, err := evtFile.ReadAt(sizeBytes, offset)

		if err == io.EOF {
			// Nothing more to read, send any pending messages and return
			sendBatch(server, batch)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error whilst reading message size from events file:%w", err)
		}

		offset += int64(read)
		msgSize := binary.BigEndian.Uint32(sizeBytes)
		msgBytes = msgBytes[:msgSize]
		read, err = evtFile.ReadAt(msgBytes, offset)
		if err != nil {
			return fmt.Errorf("error whilst reading message bytes from events file:%w", err)
		}

		offset += int64(read)

		event := &eventspb.BusEvent{}
		err = proto.Unmarshal(msgBytes, event)
		if err != nil {
			return fmt.Errorf("failed to unmarshal bus event: %w", err)
		}

		if event.Block != currentBlock {
			sendBatch(server, batch)
			time.Sleep(timeBetweenBlocks)
			currentBlock = event.Block
		}

		batch = append(batch, event)

		logEventToConsole(event)

		if len(batch) >= batchSize {
			sendBatch(server, batch)
		}
	}
}

func sendBatch(server api.CoreService_ObserveEventBusServer, batch []*eventspb.BusEvent) {
	if len(batch) > 0 {
		resp := &api.ObserveEventBusResponse{
			Events: batch,
		}
		server.SendMsg(resp)
	}
	batch = batch[:0]
}

func (e eventSource) SubmitTransaction(ctx context.Context, request *api.SubmitTransactionRequest) (*api.SubmitTransactionResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func (e eventSource) PropagateChainEvent(ctx context.Context, request *api.PropagateChainEventRequest) (*api.PropagateChainEventResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func (e eventSource) Statistics(ctx context.Context, request *api.StatisticsRequest) (*api.StatisticsResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func (e eventSource) LastBlockHeight(ctx context.Context, request *api.LastBlockHeightRequest) (*api.LastBlockHeightResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func (e eventSource) GetVegaTime(ctx context.Context, request *api.GetVegaTimeRequest) (*api.GetVegaTimeResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func (e eventSource) SubmitRawTransaction(ctx context.Context, request *api.SubmitRawTransactionRequest) (*api.SubmitRawTransactionResponse, error) {
	return nil, fmt.Errorf("not supported by eventsource tool")
}

func waitForObserveEventBusRequest(
	stream api.CoreService_ObserveEventBusServer,
) (*api.ObserveEventBusRequest, error) {
	ctx, cancelFn := context.WithCancel(stream.Context())
	oebCh := make(chan api.ObserveEventBusRequest)
	var err error
	go func() {
		defer close(oebCh)
		nb := api.ObserveEventBusRequest{}
		if err = stream.RecvMsg(&nb); err != nil {
			cancelFn()
			return
		}
		oebCh <- nb
	}()

	select {
	case <-ctx.Done():
		if err != nil {
			// this means the client disconnected
			return nil, err
		}
		return nil, ctx.Err()
	case nb := <-oebCh:
		return &nb, nil
	}
}
