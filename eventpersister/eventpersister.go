package eventpersister

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/stream"

	"github.com/golang/protobuf/proto"
)

// Run is the main function of `eventpersister` package
func Run(
	file string,
	batchSize uint,
	party, market, serverAddr, logFormat string,
	reconnect bool,
	types []string,
) error {
	flag.Parse()

	if len(serverAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	filePath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("unable to determine absolute path of file %s: %w", file, err)
	}

	fi, err := os.Create(file)
	defer fi.Close()

	if err != nil {
		return fmt.Errorf("unable to create file %s: %w", filePath, err)
	}

	fmt.Printf("persisting events to: %s\n", filePath)

	logEventToConsole, err := stream.NewLogEventToConsoleFn(logFormat)
	if err != nil {
		return err
	}

	sizeBytes := make([]byte, 4)
	handleEvent := func(e *eventspb.BusEvent) {
		size := uint32(proto.Size(e))
		protoBytes, err := proto.Marshal(e)
		if err != nil {
			panic("failed to marshal bus event:" + e.String())
		}

		binary.BigEndian.PutUint32(sizeBytes, size)
		allBytes := append(sizeBytes, protoBytes...)
		fi.Write(allBytes)
		logEventToConsole(e)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	if err := stream.ReadEvents(ctx, cancel, &wg, batchSize, party, market, serverAddr, handleEvent, reconnect, types); err != nil {
		return fmt.Errorf("error reading events: %v", err)
	}

	stream.WaitSig(ctx, cancel)
	wg.Wait()

	return nil
}
