package eventsource

import (
	"fmt"

	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"

	"github.com/golang/protobuf/proto"
	mangos "go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol"
	"go.nanomsg.org/mangos/v3/protocol/push"

	// Required for socket client
	_ "go.nanomsg.org/mangos/v3/transport/inproc"
	// Required for socket client
	_ "go.nanomsg.org/mangos/v3/transport/tcp"
)

// SocketClient stream events sent to this broker over a socket to a remote broker.
// This is used to send events from a non-validating core node to a data node.
type socketClient struct {
	address string
	sock    protocol.Socket
}

func pipeEventToString(pe mangos.PipeEvent) string {
	switch pe {
	case mangos.PipeEventAttached:
		return "Attached"
	case mangos.PipeEventDetached:
		return "Detached"
	default:
		return "Attaching"
	}
}

func newSocketClient(address string) (*socketClient, error) {
	sock, err := push.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("failed to create new push socket: %w", err)
	}

	sock.SetPipeEventHook(func(pe mangos.PipeEvent, p mangos.Pipe) {
		fmt.Println("New broker connection event")
		fmt.Printf("eventType:%s\n", pipeEventToString(pe))
		fmt.Printf("id:%d\n", p.ID())
		fmt.Printf("address:%s\n", p.Address())
	})

	s := &socketClient{
		address: address,
		sock:    sock,
	}

	return s, nil
}

func (s *socketClient) connect() error {
	err := s.sock.Dial(s.address)
	if err != nil {
		return fmt.Errorf("failed to connect to address %s: %w", s.address, err)
	}

	return nil
}

func (s *socketClient) close() {
	if err := s.sock.Close(); err != nil {
		fmt.Printf("failed to close socket:%s", err)
	}
}

func (s *socketClient) send(events []*eventspb.BusEvent) error {
	for _, event := range events {
		msg, err := proto.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		err = s.sock.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
	}
	return nil
}
