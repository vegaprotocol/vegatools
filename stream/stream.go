package stream

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"code.vegaprotocol.io/protos/vega/api"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

func connect(ctx context.Context,
	batchSize uint,
	party, market, serverAddr string) (*grpc.ClientConn, api.TradingDataService_ObserveEventBusClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	client := api.NewTradingDataServiceClient(conn)
	stream, err := client.ObserveEventBus(ctx)
	if err != nil {
		conn.Close()
		return conn, stream, err
	}

	req := &api.ObserveEventBusRequest{
		MarketId:  market,
		PartyId:   party,
		BatchSize: int64(batchSize),
		Type:      []eventspb.BusEventType{eventspb.BusEventType_BUS_EVENT_TYPE_ALL},
	}

	if err := stream.Send(req); err != nil {
		return conn, stream, fmt.Errorf("error when sending initial message in stream: %w", err)
	}
	return conn, stream, nil
}

func run(
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	batchSize uint,
	party, market, serverAddr string,
	printEvent func(string),
	reconnect bool,
) error {
	conn, stream, err := connect(ctx, batchSize, party, market, serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to event stream: %w", err)
	}

	poll := &api.ObserveEventBusRequest{
		BatchSize: int64(batchSize),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		m := jsonpb.Marshaler{}

		for true {
			defer conn.Close()
			defer stream.CloseSend()
			for {
				o, err := stream.Recv()
				if err == io.EOF {
					log.Printf("stream closed by server err=%v", err)
					break
				}
				if err != nil {
					log.Printf("stream closed err=%v", err)
					break
				}
				for _, e := range o.Events {
					estr, err := m.MarshalToString(e)
					if err != nil {
						log.Printf("unable to marshal event err=%v", err)
					}
					printEvent(estr)
				}
				if batchSize > 0 {
					if err := stream.SendMsg(poll); err != nil {
						log.Printf("failed to poll next event batch err=%v", err)
						return
					}
				}
			}

			if reconnect {
				// Keep waiting and retrying until we reconnect
				for true {
					select {
					case <-ctx.Done():
						return
					default:
						time.Sleep(time.Second * 5)
						log.Printf("Attempting to reconnect to the node")
						conn, stream, err = connect(ctx, batchSize, party, market, serverAddr)
						if err == nil {
							break
						}
					}
					if err == nil {
						break
					}
				}
			} else {
				break
			}
		}
	}()

	return nil
}

// Run is the main function of `stream` package
func Run(
	batchSize uint,
	party, market, serverAddr, logFormat string,
	reconnect bool,
) error {
	flag.Parse()

	if len(serverAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	var printEvent func(string)
	switch logFormat {
	case "raw":
		printEvent = func(event string) { fmt.Printf("%v\n", event) }
	case "text":
		printEvent = func(event string) {
			fmt.Printf("%v;%v", time.Now().UTC().Format(time.RFC3339Nano), event)
		}
	case "json":
		printEvent = func(event string) {
			fmt.Printf("{\"time\":\"%v\",%v\n", time.Now().UTC().Format(time.RFC3339Nano), event[1:])
		}
	default:
		return fmt.Errorf("error: unknown log-format: \"%v\". Allowed values: raw, text, json", logFormat)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := run(ctx, cancel, &wg, batchSize, party, market, serverAddr, printEvent, reconnect); err != nil {
		return fmt.Errorf("error when starting the stream: %v", err)
	}

	waitSig(ctx, cancel)
	wg.Wait()

	return nil
}

func waitSig(ctx context.Context, cancel func()) {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	select {
	case sig := <-gracefulStop:
		log.Printf("Caught signal name=%v", sig)
		log.Printf("closing client connections")
		cancel()
	case <-ctx.Done():
		return
	}
}
