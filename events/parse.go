package events

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func Run(evtIn, JSONin, out string, num uint64, types []int32, create bool) error {
	evtTypes := map[eventspb.BusEventType]struct{}{}
	// remove unspecified / all from input
	for i := 0; i < len(types); i++ {
		v := types[i]
		if v == 0 || v == 1 {
			types = append(types[:i], types[i+1:]...)
			i--
		}
	}
	if len(types) == 0 { // emtpy == default to all
		for k := range eventspb.BusEventType_name {
			if k == 0 || k == 1 { // skip unspecified || all
				continue
			}
			types = append(types, k)
		}
	}
	// set filter type map
	for _, v := range types {
		evtTypes[eventspb.BusEventType(v)] = struct{}{}
	}
	if len(evtIn) != 0 {
		return evtToJSON(evtIn, out, evtTypes, num)
	}
	if !create {
		return JSONToEvt(JSONin, os.Stdout, evtTypes, num)
	}
	outF, err := os.Create(out)
	if err != nil {
		return err
	}
	return JSONToEvt(JSONin, outF, evtTypes, num)
}

func JSONToEvt(jsonIn string, out *os.File, types map[eventspb.BusEventType]struct{}, num uint64) error {
	in, err := os.Open(jsonIn)
	if err != nil {
		return err
	}
	defer in.Close()
	cnt := uint64(0)
	unmarshaler := jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	dec := json.NewDecoder(in)
	size := make([]byte, 4)
	for {
		// we have a JSON array, so read the opening bracket token
		// or the comma after each element
		e := &eventspb.BusEvent{}
		err := unmarshaler.UnmarshalNext(dec, e)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if _, ok := types[e.Type]; !ok {
			continue
		}
		data, err := proto.Marshal(e)
		if err != nil {
			return err
		}
		binary.BigEndian.PutUint32(size, uint32(len(data)))
		if _, err := out.Write(append(size, data...)); err != nil {
			return err
		}
		cnt++
		if num > 0 && cnt >= num {
			return nil
		}
	}
}

func evtToJSON(evtIn, out string, types map[eventspb.BusEventType]struct{}, num uint64) error {
	ctx, cfunc := context.WithCancel(context.Background())
	outF := os.Stdout
	if len(out) != 0 {
		of, err := os.Create(out)
		if err != nil {
			cfunc()
			return err
		}
		defer of.Close()
		outF = of
	}
	defer cfunc()
	ch, ech := startFileRead(ctx, evtIn)
	parsed := uint64(0)
	marshaler := jsonpb.Marshaler{
		EnumsAsInts: true,
		OrigName:    true,
		Indent:      "   ",
	}
	nl := "\n"
	for {
		select {
		case e, ok := <-ch:
			if e == nil && !ok {
				return nil
			}
			if _, ok := types[e.Type]; !ok {
				continue
			}
			es, err := marshaler.MarshalToString(e)
			if err != nil {
				return err
			}
			parsed++
			if _, err := outF.WriteString(es + nl); err != nil {
				return err
			}
			if num > 0 && parsed >= num {
				return nil
			}
		case err, ok := <-ech:
			if err == nil && !ok {
				return nil
			}
			return err
		}
	}
}

func readRawEvent(eventFile *os.File, offset int64) (event []byte, seqNum uint64,
	totalBytesRead uint32, err error,
) {
	sizeBytes := make([]byte, 4)
	read, err := eventFile.ReadAt(sizeBytes, offset)

	if err == io.EOF {
		return nil, 0, 0, nil
	} else if err != nil {
		return nil, 0, 0, fmt.Errorf("error reading message size from events file:%w", err)
	}

	if read < 4 {
		return nil, 0, 0, nil
	}

	messageOffset := offset + 4

	msgSize := binary.BigEndian.Uint32(sizeBytes)
	seqNumAndMsgBytes := make([]byte, msgSize)
	read, err = eventFile.ReadAt(seqNumAndMsgBytes, messageOffset)
	if err == io.EOF {
		return nil, 0, 0, nil
	} else if err != nil {
		return nil, 0, 0, fmt.Errorf("error reading message bytes from events file:%w", err)
	}

	if read < int(msgSize) {
		return nil, 0, 0, nil
	}

	seqNumBytes := seqNumAndMsgBytes[:8]
	seqNum = binary.BigEndian.Uint64(seqNumBytes)
	msgBytes := seqNumAndMsgBytes[8:]
	totalBytesRead = 4 + msgSize

	return msgBytes, seqNum, totalBytesRead, nil
}

func startFileRead(ctx context.Context, file string) (<-chan *eventspb.BusEvent, <-chan error) {
	ch := make(chan *eventspb.BusEvent, 1)
	ech := make(chan error, 1)
	eventFile, err := os.Open(file)
	if err != nil {
		ech <- err
		defer func() {
			close(ch)
			close(ech)
		}()
		return ch, ech
	}
	go func() {
		// close and cleanup everything once we're done
		defer func() {
			eventFile.Close()
			close(ch)
			close(ech)
		}()

		// sizeBytes := make([]byte, 4)
		var offset int64 = 0

		// read the input file and push everything onto a channel
		for {
			select {
			case <-ctx.Done():
				return
			default:
				rawEvent, _, read, err := readRawEvent(eventFile, offset)
				if err != nil {
					return
				}

				if read == 0 {
					return
				}

				offset += int64(read)

				// We have to deserialize the busEvent here (even though we output the raw busEvent)
				// to be able to skip the first few events before we get a BeginBlock and to be
				// able to sleep between blocks.
				busEvent := &eventspb.BusEvent{}
				if err := proto.Unmarshal(rawEvent, busEvent); err != nil {
					ech <- fmt.Errorf("failed to unmarshal bus event: %w", err)
				} else {
					ch <- busEvent
				}
			}
		}
	}()
	return ch, ech
}
