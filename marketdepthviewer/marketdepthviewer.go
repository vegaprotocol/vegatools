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

	api "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	proto "code.vegaprotocol.io/vega/protos/vega"
	eventspb "code.vegaprotocol.io/vega/protos/vega/events/v1"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

// Opts command line options
type Opts struct {
	Market     string
	ServerAddr string
	UseDeltas  bool
}

// MarketDepthBook structure to hold market depth for delta updates
type MarketDepthBook struct {
	buys   map[string]*proto.PriceLevel
	sells  map[string]*proto.PriceLevel
	seqNum uint64
}

type mdv struct {
	ts           tcell.Screen
	redStyle     tcell.Style
	greenStyle   tcell.Style
	whiteStyle   tcell.Style
	market       *proto.Market
	marketData   *proto.MarketData
	updateMode   string
	displayMutex sync.Mutex

	// Variables to control drawing speed
	lastRedraw time.Time
	dirty      bool

	mode proto.Market_TradingMode

	book MarketDepthBook
}

func (m *mdv) getMarketToDisplay(dataclient api.TradingDataServiceClient, marketID string) (*proto.Market, error) {
	marketsRequest := &api.ListMarketsRequest{}

	marketsResponse, err := dataclient.ListMarkets(context.Background(), marketsRequest)
	if err != nil {
		return nil, err
	}

	// If the user has picked a market already that is valid, use that
	for _, market := range marketsResponse.Markets.Edges {
		if market.Node.Id == marketID {
			return market.Node, nil
		}
	}

	// If we have no markets, lets quit now
	if len(marketsResponse.Markets.Edges) == 0 {
		return nil, nil
	}

	// If there is only one market, pick that automatically
	if len(marketsResponse.Markets.Edges) == 1 {
		return marketsResponse.Markets.Edges[0].Node, nil
	}

	// Print out all the markets with their index
	for index, market := range marketsResponse.Markets.Edges {
		fmt.Printf("[%d]:%s (%s) [%s]\n", index+1, market.Node.State.String(), market.Node.TradableInstrument.Instrument.Name, market.Node.Id)
	}

	// Ask the user to select a market
	fmt.Printf("Which market do you want to view? ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read option:", err)
		return nil, err
	}

	// Convert input into an index
	input = strings.Replace(input, "\n", "", -1)
	input = strings.Replace(input, "\r", "", -1)
	index, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Failed to convert input into index:", err)
		return nil, err
	}
	// Correct the index value as we start with 0
	index--

	if index < 0 || index > len(marketsResponse.Markets.Edges)-1 {
		fmt.Println("Invalid market selected")
		return nil, fmt.Errorf("invalid market selection: %s", marketID)
	}

	fmt.Println("Using market:", index)

	return marketsResponse.Markets.Edges[index].Node, nil
}

func (m *mdv) subscribeToMarketData(dataclient api.TradingDataServiceClient) error {
	events := []eventspb.BusEventType{eventspb.BusEventType_BUS_EVENT_TYPE_MARKET_DATA}
	eventBusDataReq := &api.ObserveEventBusRequest{
		Type:     events,
		MarketId: m.market.Id,
	}

	stream, err := dataclient.ObserveEventBus(context.Background())
	if err != nil {
		return fmt.Errorf("failed to subscribe to event bus data: %w", err)
	}

	// Then we subscribe to the data
	err = stream.SendMsg(eventBusDataReq)
	if err != nil {
		return fmt.Errorf("unable to send event bus request on the stream: %w", err)
	}
	go m.processMarketDataSubscription(stream)
	return err
}

func (m *mdv) processMarketDataSubscription(stream api.TradingDataService_ObserveEventBusClient) {
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
				m.mode = event.GetMarketData().MarketTradingMode
				m.marketData = event.GetMarketData()
			}
		}
	}
}

func (m *mdv) subscribeMarketDepthSnapshots(dataclient api.TradingDataServiceClient) error {
	req := &api.ObserveMarketsDepthRequest{
		MarketIds: []string{m.market.Id},
	}
	stream, err := dataclient.ObserveMarketsDepth(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to trades: %w", err)
	}

	m.ts.Clear()
	m.drawHeaders()
	m.drawTime()
	m.ts.Show()

	// Run in background and process messages
	go m.processMarketDepth(stream)
	return nil
}

func (m *mdv) processMarketDepth(stream api.TradingDataService_ObserveMarketsDepthClient) {
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
		w, h := m.ts.Size()

		m.ts.Clear()
		m.drawHeaders()
		m.drawTime()
		m.drawMarketState()

		if len(o.MarketDepth) > 0 {
			// Only use the latest depth snapshot
			md := o.MarketDepth[len(o.MarketDepth)-1]
			m.drawSequenceNumber(md.SequenceNumber)

			var bidVolume uint64
			var askVolume uint64

			// Print Buys
			buyPriceLevels := md.Buy
			for index, pl := range buyPriceLevels {
				bidVolume += pl.Volume
				if index > (h - 6) {
					continue
				}
				text := fmt.Sprintf("%12d", pl.Volume)
				m.drawString((w/4)-21, index+4, m.greenStyle, text)
				text = fmt.Sprintf("%12s", pl.Price)
				m.drawString((w/4)+7, index+4, m.greenStyle, text)
			}

			// Print Sells
			sellPriceLevels := md.Sell
			for index, pl := range sellPriceLevels {
				askVolume += pl.Volume
				if index > (h - 6) {
					continue
				}
				m.drawString((3*w/4)-22, index+4, m.redStyle, pl.Price)
				text := fmt.Sprintf("%d", pl.Volume)
				m.drawString((3*w/4)+9, index+4, m.redStyle, text)
			}

			text := fmt.Sprintf("%8d", bidVolume)
			m.drawString((w / 4), h-1, m.whiteStyle, text)
			text = fmt.Sprintf("%8d", askVolume)
			m.drawString((3 * w / 4), h-1, m.whiteStyle, text)

			m.ts.Show()
		}
	}
}

// Run is the main entry point for this tool
func Run(opts Opts) error {
	m := mdv{book: MarketDepthBook{buys: map[string]*proto.PriceLevel{}, sells: map[string]*proto.PriceLevel{}}}
	// Create connection to vega
	connection, err := grpc.Dial(opts.ServerAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("failed to connect to the vega gRPC port: %s", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	// Look up all the markets on this node
	m.market, err = m.getMarketToDisplay(dataclient, opts.Market)
	if err != nil {
		return err
	}

	if m.market == nil {
		return fmt.Errorf("failed to get market details")
	}

	err = m.initialiseScreen()
	if err != nil {
		return err
	}

	// Subscribe to the market stream to listen for market state
	err = m.subscribeToMarketData(dataclient)
	if err != nil {
		return err
	}

	// Make the decision here if we are using snapshots or deltas
	if opts.UseDeltas {
		m.updateMode = "(DELTAS)"
		// Using deltas to update a snapshot
		err = m.subscribeToMarketDepthUpdates(dataclient)
		if err != nil {
			return err
		}
		// Get one snapshot to act as the base
		err = m.getMarketDepthSnapshot(dataclient)
		if err != nil {
			return err
		}
	} else {
		m.updateMode = "(SNAPSHOTS)"
		// Getting regular snapshots
		err = m.subscribeMarketDepthSnapshots(dataclient)
		if err != nil {
			return err
		}
	}

	for {
		switch ev := m.ts.PollEvent().(type) {
		case *tcell.EventResize:
			m.ts.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Rune() == 'q' {
				m.ts.Fini()
				os.Exit(0)
			}
		}
	}
}
