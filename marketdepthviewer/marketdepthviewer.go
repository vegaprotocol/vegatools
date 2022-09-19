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
	"sync"
	"time"

	api "code.vegaprotocol.io/vega/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/vega/protos/vega"
	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

type Opts struct {
	Market     string
	ServerAddr string
	UseDeltas  bool
}

type MarketDepthBook struct {
	buys   map[string]*proto.PriceLevel
	sells  map[string]*proto.PriceLevel
	seqNum uint64
}

var (
	ts              tcell.Screen
	redStyle        tcell.Style
	greenStyle      tcell.Style
	whiteStyle      tcell.Style
	inverseRedStyle tcell.Style
	market          *proto.Market
	marketData      *proto.MarketData
	updateMode      string
	displayMutex    sync.Mutex

	// Variables to control drawing speed
	lastRedraw time.Time
	dirty      bool

	mode proto.Market_TradingMode

	book MarketDepthBook = MarketDepthBook{buys: map[string]*proto.PriceLevel{}, sells: map[string]*proto.PriceLevel{}}
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

	// If we have no markets, lets quit now
	if len(marketsResponse.Markets) == 0 {
		return nil
	}

	// If there is only one market, pick that automatically
	if len(marketsResponse.Markets) == 1 {
		return marketsResponse.Markets[0]
	}

	// Print out all the markets with their index
	for index, market := range marketsResponse.Markets {
		fmt.Printf("[%d]:%s (%s) [%s]\n", index+1, market.State.String(), market.TradableInstrument.Instrument.Name, market.Id)
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
	input = strings.Replace(input, "\r", "", -1)
	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Failed to convert input into index:", err)
		os.Exit(0)
	}
	// Correct the index value as we start with 0
	index--

	if index < 0 || index > len(marketsResponse.Markets)-1 {
		fmt.Println("Invalid market selected")
		os.Exit(0)
	}

	fmt.Println("Using market:", index)

	return marketsResponse.Markets[index]
}

func subscribeToMarketData(dataclient api.TradingDataServiceClient) {
	events := []eventspb.BusEventType{eventspb.BusEventType_BUS_EVENT_TYPE_MARKET_DATA}
	eventBusDataReq := &api.ObserveEventBusRequest{
		Type:     events,
		MarketId: market.Id,
	}

	stream, err := dataclient.ObserveEventBus(context.Background())
	if err != nil {
		log.Panicln("Failed to subscribe to event bus data: ", err)
	}

	// Then we subscribe to the data
	err = stream.SendMsg(eventBusDataReq)
	if err != nil {
		log.Panicln("Unable to send event bus request on the stream", err)
	}
	go processMarketDataSubscription(stream)
}

func processMarketDataSubscription(stream api.TradingDataService_ObserveEventBusClient) {
	for {
		eb, err := stream.Recv()
		if err == io.EOF {
			log.Panicln("event bus data: stream closed by server err:", err)
			break
		}
		if err != nil {
			log.Panicln("event bus data: stream closed err:", err)
			break
		}

		for _, event := range eb.Events {
			switch event.Type {
			case eventspb.BusEventType_BUS_EVENT_TYPE_MARKET_DATA:
				mode = event.GetMarketData().MarketTradingMode
				marketData = event.GetMarketData()
			}
		}
	}
}

func subscribeMarketDepthSnapshots(dataclient api.TradingDataServiceClient) {
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
	go processMarketDepth(stream)
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
		drawMarketState()

		var bidVolume uint64
		var askVolume uint64

		// Print Buys
		buyPriceLevels := o.MarketDepth.Buy
		for index, pl := range buyPriceLevels {
			bidVolume += pl.Volume
			if index > (h - 6) {
				continue
			}
			text := fmt.Sprintf("%12d", pl.Volume)
			drawString((w/4)-21, index+4, greenStyle, text)
			text = fmt.Sprintf("%12s", pl.Price)
			drawString((w/4)+7, index+4, greenStyle, text)
		}

		// Print Sells
		sellPriceLevels := o.MarketDepth.Sell
		for index, pl := range sellPriceLevels {
			askVolume += pl.Volume
			if index > (h - 6) {
				continue
			}
			text := fmt.Sprintf("%s", pl.Price)
			drawString((3*w/4)-22, index+4, redStyle, text)
			text = fmt.Sprintf("%d", pl.Volume)
			drawString((3*w/4)+9, index+4, redStyle, text)
		}

		text := fmt.Sprintf("%8d", bidVolume)
		drawString((w / 4), h-1, whiteStyle, text)
		text = fmt.Sprintf("%8d", askVolume)
		drawString((3 * w / 4), h-1, whiteStyle, text)

		ts.Show()
	}
}

// Run is the main entry point for this tool
func Run(opts Opts) error {
	// Create connection to vega
	connection, err := grpc.Dial(opts.ServerAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("failed to connect to the vega gRPC port: %s", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	// Look up all the markets on this node
	market = getMarketToDisplay(dataclient, opts.Market)
	if market == nil {
		return fmt.Errorf("failed to get market details")
	}

	initialiseScreen()

	// Subscribe to the market stream to listen for market state
	subscribeToMarketData(dataclient)

	// Make the decision here if we are using snapshots or deltas
	if opts.UseDeltas {
		updateMode = "(DELTAS)"
		// Using deltas to update a snapshot
		subscribeToMarketDepthUpdates(dataclient)
		// Get one snapshot to act as the base
		err = getMarketDepthSnapshot(dataclient)
		if err != nil {
			return err
		}
	} else {
		updateMode = "(SNAPSHOTS)"
		// Getting regular snapshots
		subscribeMarketDepthSnapshots(dataclient)
	}

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
