package delegationviewer

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	api "code.vegaprotocol.io/vega/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/vega/protos/vega"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

var (
	ts         tcell.Screen
	redStyle   tcell.Style
	greenStyle tcell.Style
	whiteStyle tcell.Style
	greyStyle  tcell.Style
	nodes      []*proto.Node

	args struct {
		gRPCAddress string
		updateDelay uint
	}

	validators map[string]string
)

func initialiseValidatorNames() {
	validators = map[string]string{
		"126751c5830b50d39eb85412fb2964f46338cce6946ff455b73f1b1be3f5e8cc": "Greenfield One",
		"25794776055552a92e7b27dd8f15563ffb78defe7694d6c4da8bb258daca897c": "Lovali",
		"43697a3e911d8b70c0ce672adde17a5c38ca8f6a0486bf85ed0546e1b9a82887": "B-Harvest",
		"4f69b1784656174e89eb094513b7136e88670b42517ed0e48cb6fd3062eb8478": "Nodes Guru",
		"55504e9bfd914a7bbefa342c82f59a2f4dee344e5b6863a14c02a812f4fbde32": "RBF",
		"5ca98e0dd81143fafea3a3abcefafee73f3886ac97053db8b446593e75c10e9d": "P2P.ORG",
		"5db9794f44c85b4b259907a00c8ea2383ad688dfef6ffb72c8743b6ae3eaefd4": "Ryabina",
		"74023df02b8afc9eaf3e3e2e8b07eab1d2122ac3e74b1b0222daf4af565ad3dd": "XPRV",
		"8d33c6e06207ed5735c8b5b6c0c6234f44eb381b242a25a538ed3315369d2203": "Nala DAO",
		"ac735acc9ab11cf1d8c59c2df2107e00092b4ac96451cb137a1629af5b66242a": "Figment",
		"b861c11eb825d55f835aec898b3caae66a681a354bcb59651d5b3faf02b34844": "Commodum DAO",
		"efbdf943443bd7595e83b0d7e88f37b7932d487d1b94aab3d004997273bb43fc": "Chorus One",
		"f3022974212780ea1196af08fd2e8a9c0d784d0be8e97637bd5e763ac4c219bd": "Staking Facilities",
	}
}

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
		err := getDelegationDetails(dataclient)
		if err != nil {
			log.Fatalln("Failed to get node information", err)
		}
		ts.Clear()
		drawHeaders()
		drawTime()
		drawNodes()
		ts.Show()
		time.Sleep(time.Second * time.Duration(args.updateDelay))
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
	greyStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorLightGrey)
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
func drawStringPc(x, y, f int, style tcell.Style, str string) {
	w, _ := ts.Size()
	if x > 0 {
		x = (w * x) / 100
	}

	// See if we need to left pad
	offset := 0
	if len(str) < f {
		offset = f - len(str)
	}

	drawString(x+offset, y, style, str)
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
	drawString((w/2)-12, 0, whiteStyle, "Delegation Details")
	drawStringPc(0, 2, 0, whiteStyle, "Node Id")

	drawStringPc(45, 1, 0, whiteStyle, "Staked By")
	drawStringPc(25, 2, 0, whiteStyle, " Operator")
	drawStringPc(40, 2, 0, whiteStyle, "Delegates")
	drawStringPc(55, 2, 0, whiteStyle, "    Total")
	drawStringPc(70, 2, 0, whiteStyle, "  Pending")
	drawStringPc(85, 2, 0, whiteStyle, "   Status")
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
		if i%2 == 0 {
			style = greyStyle
		}

		name, ok := validators[node.Id]
		if !ok {
			name = node.Id
		}

		drawStringPc(0, 3+i, 0, style, name)
		drawStringPc(25, 3+i, 9, style, scaleValues(node.StakedByOperator))
		drawStringPc(40, 3+i, 9, style, scaleValues(node.StakedByDelegates))
		drawStringPc(55, 3+i, 9, style, scaleValues(node.StakedTotal))

		if strings.Contains(node.PendingStake, "-") {
			drawStringPc(70, 3+i, 9, redStyle, scaleValues(node.PendingStake))
		} else {
			drawStringPc(70, 3+i, 9, greenStyle, scaleValues(node.PendingStake))
		}

		// We remove the "NODE_STATUS_" part of the string
		status := node.Status.String()
		drawStringPc(85, 3+i, 0, whiteStyle, status[12:])
	}
}

func scaleValues(amount string) string {
	if len(amount) > 18 {
		return amount[:len(amount)-18]
	}
	return amount
}

// Run is the main entry point for this tool
func Run(gRPCAddress string, delay uint) error {
	args.gRPCAddress = gRPCAddress
	args.updateDelay = delay

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

	initialiseValidatorNames()

	initialiseScreen()

	// Start a thread to poll for delegation details
	go updateDelegationDetails(dataclient)

	for {
		switch ev := ts.PollEvent().(type) {
		case *tcell.EventResize:
			ts.Sync()
			ts.Clear()
			drawHeaders()
			drawTime()
			drawNodes()
			ts.Show()

		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Rune() == 'q' {
				ts.Fini()
				os.Exit(0)
			}
		}
	}
}
