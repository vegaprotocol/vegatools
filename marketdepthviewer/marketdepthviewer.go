package marketdepthviewer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vegaprotocol/api/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api/go/generated/code.vegaprotocol.io/vega/proto/api"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

var (
	ts         tcell.Screen
	redStyle   tcell.Style
	greenStyle tcell.Style
	whiteStyle tcell.Style
	market     *proto.Market

	args struct {
		gRPCAddress string
		marketID    string
	}
)

func getMarketToDisplay(dataclient api.TradingDataServiceClient, marketID string) *proto.Market {
	marketsRequest := &api.MarketsRequest{}

	marketsResponse, err := dataclient.Markets(context.Background(), marketsRequest)
	if err != nil {
		return nil
	}

	// If the user has picked a market already that is valid, use that
	for _, market := range marketsResponse.Markets {
		if market.Id == marketID {
			return market
		}
	}

	// Print out all the markets with their index
	for index, market := range marketsResponse.Markets {
		fmt.Printf("[%d]:%s (%s) [%s]\n", index, market.State.String(), market.TradableInstrument.Instrument.Name, market.Id)
	}

	// Ask the user to select a market
	fmt.Printf("Which market do you want to view? ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read option:", err)
		os.Exit(0)
	}

	// Convert input into an index
	input = strings.Replace(input, "\n", "", -1)
	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Failed to convert input into index:", err)
		os.Exit(0)
	}

	if index < 0 || index > len(marketsResponse.Markets)-1 {
		fmt.Println("Invalid market selected")
		os.Exit(0)
	}

	fmt.Println("Using market:", index)

	return marketsResponse.Markets[index]
}

func getMarketDepth(dataclient api.TradingDataServiceClient) {
	req := &api.MarketDepthSubscribeRequest{
		MarketId: market.Id,
	}
	stream, err := dataclient.MarketDepthSubscribe(context.Background(), req)
	if err != nil {
		log.Fatalln("Failed to subscribe to trades: ", err)
		os.Exit(0)
	}

	ts.Clear()
	drawHeaders()
	drawTime()
	ts.Show()

	// Run in background and process messages
	processMarketDepth(stream)
}

func getMarketDepthUpdates(dataclient api.TradingDataServiceClient) {

	req := &api.MarketDepthUpdatesSubscribeRequest{
		MarketId: market.Id,
	}
	stream, err := dataclient.MarketDepthUpdatesSubscribe(context.Background(), req)
	if err != nil {
		log.Fatalln("Failed to subscribe to trades: ", err)
	}

	ts.Clear()
	drawHeaders()
	drawTime()
	ts.Show()

	// Run in background and process messages
	processMarketDepthUpdates(stream)
}

func processMarketDepthUpdates(stream api.TradingDataService_MarketDepthUpdatesSubscribeClient) {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			log.Println("orders: stream closed by server err: ", err)
			break
		}
		if err != nil {
			log.Println("orders: stream closed err: ", err)
			break
		}
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

func drawString(x, y int, style tcell.Style, str string) {
	for i, c := range str {
		ts.SetContent(x+i, y, c, nil, style)
	}
}

func drawHeaders() {
	w, h := ts.Size()

	// Draw the headings
	drawString((w/4)-2, 1, whiteStyle, "Bids")
	drawString((3*w/4)-2, 1, whiteStyle, "Asks")

	drawString((w/4)-19, 2, whiteStyle, "--Volume--")
	drawString((w/4)+8, 2, whiteStyle, "---Price---")
	drawString((3*w/4)-22, 2, whiteStyle, "---Price---")
	drawString((3*w/4)+9, 2, whiteStyle, "--Volume--")

	// If we have a market name, use that
	if market != nil {
		text := fmt.Sprintf("Market: %s [%s]", market.TradableInstrument.Instrument.Name, market.Id)
		drawString(0, 0, whiteStyle, text)
	} else {
		text := fmt.Sprintf("Market: %s", market.Id)
		drawString(0, 0, whiteStyle, text)
	}
	drawString(w-26, 0, whiteStyle, "Last Update Time:")

	drawString((w/4)-8, h-1, whiteStyle, "Volume:")
	drawString((3*w/4)-8, h-1, whiteStyle, "Volume:")
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}

func drawSequenceNumber(seqNum uint64) {
	w, _ := ts.Size()
	text := fmt.Sprintf("SeqNum:%6d", seqNum)
	drawString((w/2)-6, 0, whiteStyle, text)
}

func processMarketDepth(stream api.TradingDataService_MarketDepthSubscribeClient) {
	for {
		o, err := stream.Recv()
		if err == io.EOF {
			log.Println("orders: stream closed by server err: ", err)
			break
		}
		if err != nil {
			log.Println("orders: stream closed err: ", err)
			break
		}

		w, h := ts.Size()

		ts.Clear()
		drawHeaders()
		drawTime()
		drawSequenceNumber(o.MarketDepth.SequenceNumber)

		var bidVolume uint64 = 0
		var askVolume uint64 = 0

		// Print Buys
		buyPriceLevels := o.MarketDepth.Buy
		for index, pl := range buyPriceLevels {
			bidVolume += pl.Volume
			if index > (h - 6) {
				continue
			}
			text := fmt.Sprintf("%12d", pl.Volume)
			drawString((w/4)-21, index+3, greenStyle, text)
			text = fmt.Sprintf("%12d", pl.Price)
			drawString((w/4)+7, index+3, greenStyle, text)
		}

		// Print Sells
		sellPriceLevels := o.MarketDepth.Sell
		for index, pl := range sellPriceLevels {
			askVolume += pl.Volume
			if index > (h - 6) {
				continue
			}
			text := fmt.Sprintf("%d", pl.Price)
			drawString((3*w/4)-22, index+3, redStyle, text)
			text = fmt.Sprintf("%d", pl.Volume)
			drawString((3*w/4)+9, index+3, redStyle, text)
		}

		text := fmt.Sprintf("%8d", bidVolume)
		drawString((w / 4), h-1, whiteStyle, text)
		text = fmt.Sprintf("%8d", askVolume)
		drawString((3 * w / 4), h-1, whiteStyle, text)

		ts.Show()
	}
}

// Run is the main entry point for this tool
func Run(gRPCAddress, marketID string) error {
	// Create connection to vega
	connection, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("Failed to connect to the vega gRPC port: %s", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	// Look up all the markets on this node
	market = getMarketToDisplay(dataclient, marketID)
	if market == nil {
		return fmt.Errorf("Failed to get market details")
	}

	initialiseScreen()

	// Start the book displaying
	go getMarketDepth(dataclient)

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
