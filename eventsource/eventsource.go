package eventsource

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"

	"github.com/golang/protobuf/proto"
)

func sendAllEvents(sendEvents func([]*eventspb.BusEvent) error, evtFile *os.File,
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
			// Nothing more to read, sendEvents any pending messages and return
			return sendBatch(sendEvents, batch)
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
			err = sendBatch(sendEvents, batch)
			batch = batch[:0]
			if err != nil {
				return err
			}
			time.Sleep(timeBetweenBlocks)
			currentBlock = event.Block
		}

		batch = append(batch, event)

		logEventToConsole(event)

		if len(batch) >= batchSize {
			err = sendBatch(sendEvents, batch)
			batch = batch[:0]
			if err != nil {
				return err
			}
		}
	}
}

func sendBatch(sendEvents func([]*eventspb.BusEvent) error, batch []*eventspb.BusEvent) error {
	if len(batch) > 0 {
		err := sendEvents(batch)
		if err != nil {
			return fmt.Errorf("failed to send batch: %w", err)
		}
	}
	return nil
}
