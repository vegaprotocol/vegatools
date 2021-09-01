package delegationviewer

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	api "code.vegaprotocol.io/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/protos/vega"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

var (
	ts         tcell.Screen
	redStyle   tcell.Style
	greenStyle tcell.Style
	whiteStyle tcell.Style
	nodes      []*proto.Node

	args struct {
		gRPCAddress string
	}
)

func getDelegationDetails(dataclient api.TradingDataServiceClient) error {
	req := &api.GetNodesRequest{}
	nodeResp, err := dataclient.GetNodes(context.Background(), req)
	if err != nil {
		return fmt.Errorf("Failed to get node details: %v", err)
	}
	nodes = nodeResp.Nodes
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Id < nodes[j].Id })
	return nil
}

func updateDelegationDetails(dataclient api.TradingDataServiceClient) {
	for {
		time.Sleep(time.Second * 3)
		err := getDelegationDetails(dataclient)
		if err != nil {
			log.Fatalln("Failed to get node information", err)
		}
		ts.Clear()
		drawHeaders()
		drawTime()
		drawNodes()
		ts.Show()
	}
}

func initialiseScreen() error {
	var e error
	ts, e = tcell.NewScreen()
	if e != nil {
		log.Fatalln("Failed to create new tcell screen", e)
		return e
	}

	e = ts.Init()
	if e != nil {
		log.Fatalln("Failed to initialise the tcell screen", e)
		return e
	}

	whiteStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorWhite)
	greenStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorGreen)
	redStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorRed)

	return nil
}

// Draws a string starting at the x percentage of the column
// e.g 0% starts on the left, 50% starts half way across
func drawStringPc(x, y int, style tcell.Style, str string) {
	w, _ := ts.Size()
	if x > 0 {
		x = (w * x) / 100
	}
	drawString(x, y, style, str)
}

func drawString(x, y int, style tcell.Style, str string) {
	for i, c := range str {
		ts.SetContent(x+i, y, c, nil, style)
	}
}

func drawHeaders() {
	w, _ := ts.Size()

	// Draw the headings
	drawString(0, 0, whiteStyle, args.gRPCAddress)
	drawString((w/2)-9, 0, whiteStyle, "Delegation Details")
	drawStringPc(0, 2, whiteStyle, "Id")
	drawStringPc(30, 2, whiteStyle, "Location")

	drawStringPc(55, 1, whiteStyle, "Staked By")
	drawStringPc(50, 2, whiteStyle, "Operator")
	drawStringPc(60, 2, whiteStyle, "Delegates")
	drawStringPc(70, 2, whiteStyle, "MaxIntended")
	drawStringPc(80, 2, whiteStyle, "Pending")
	drawStringPc(90, 2, whiteStyle, "Status")
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}

func drawNodes() {
	for i, node := range nodes {
		style := whiteStyle
		switch node.Status {
		case proto.NodeStatus_NODE_STATUS_VALIDATOR:
			style = whiteStyle
		case proto.NodeStatus_NODE_STATUS_NON_VALIDATOR:
			style = greenStyle
		case proto.NodeStatus_NODE_STATUS_UNSPECIFIED:
			style = redStyle
		}

		drawStringPc(0, 3+i, style, node.Id)
		drawStringPc(30, 3+i, style, node.Location)
		drawStringPc(50, 3+i, style, node.StakedByOperator)
		drawStringPc(60, 3+i, style, node.StakedByDelegates)
		drawStringPc(70, 3+i, style, node.MaxIntendedStake)
		drawStringPc(80, 3+i, style, node.PendingStake)

		// We remove the "NODE_STATUS_" part of the string
		status := node.Status.String()
		drawStringPc(90, 3+i, whiteStyle, status[12:])
	}
}

// Run is the main entry point for this tool
func Run(gRPCAddress string) error {
	args.gRPCAddress = gRPCAddress
	// Create connection to vega
	connection, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("Failed to connect to the vega gRPC port: %s", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	// Check we can get delegation information
	err = getDelegationDetails(dataclient)
	if err != nil {
		return fmt.Errorf("Failed to get delegation details: %v", err)
	}

	initialiseScreen()

	// Start a thread to poll for delegation details
	go updateDelegationDetails(dataclient)

	for {
		switch ev := ts.PollEvent().(type) {
		case *tcell.EventResize:
			ts.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Rune() == 'q' {
				ts.Fini()
				os.Exit(0)
			}
		}
	}
}
