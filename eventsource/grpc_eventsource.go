package eventsource

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	api "code.vegaprotocol.io/protos/vega/api/v1"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"

	"google.golang.org/grpc"
)

// RunGrpcEventSource is the main function of `eventsource` package
func RunGrpcEventSource(eventsFile string, port uint, closeConnection bool, timeBetweenBlocks time.Duration, consoleLogFormat string) error {
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

	return grpcServer.Serve(listener)
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
	fmt.Printf("creating event source for events at: %s\n", filePath)

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

	sendEvents := func(batch []*eventspb.BusEvent) error {
		resp := &api.ObserveEventBusResponse{
			Events: batch,
		}
		return server.SendMsg(resp)
	}

	err = sendAllEvents(sendEvents, fi, int(req.BatchSize), e.timeBetweenBlocks, e.logEventToConsole)

	if !e.closeConnection {
		for {
			time.Sleep(time.Second)
		}
	}

	return nil
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
