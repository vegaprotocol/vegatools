package liquiditycommitment

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	api "code.vegaprotocol.io/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/protos/vega"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

var (
	ts             tcell.Screen
	greyStyle      tcell.Style
	whiteStyle     tcell.Style
	market         *proto.Market
	mapMarketToLPs map[string][]*proto.LiquidityProvision = map[string][]*proto.LiquidityProvision{}
	partyToLps     map[string]*proto.LiquidityProvision   = map[string]*proto.LiquidityProvision{}
	marketData     *proto.MarketData
)

func getLiquidityProvisions(dataclient api.TradingDataServiceClient, marketID string) []*proto.LiquidityProvision {
	lpReq := &api.LiquidityProvisionsRequest{Market: marketID}

	response, err := dataclient.LiquidityProvisions(context.Background(), lpReq)
	if err != nil {
		log.Println(err)
	}
	return response.LiquidityProvisions
}

func getMarketData(dataclient api.TradingDataServiceClient, marketID string) *proto.MarketData {
	fmt.Println("marketID=", marketID)
	marketDataRequest := &api.MarketDataByIDRequest{
		MarketId: marketID,
	}

	marketDataResponse, err := dataclient.MarketDataByID(context.Background(), marketDataRequest)
	if err != nil {
		fmt.Println("Failed to get market data")
		os.Exit(0)
		return nil
	}
	return marketDataResponse.MarketData
}

func getMarketToDisplay(dataclient api.TradingDataServiceClient, marketID string) *proto.Market {
	marketsRequest := &api.MarketsRequest{}

	marketsResponse, err := dataclient.Markets(context.Background(), marketsRequest)
	if err != nil {
		return nil
	}

	var validMarkets []*proto.Market
	// Check each market to see if we have at least one LP
	for _, market := range marketsResponse.Markets {
		lps := getLiquidityProvisions(dataclient, market.Id)
		if len(lps) > 0 {
			validMarkets = append(validMarkets, market)
			mapMarketToLPs[market.Id] = lps
		}
	}

	if len(validMarkets) == 0 {
		return nil
	}

	// If the user has picked a market already that is valid, use that
	for _, market := range validMarkets {
		if market.Id == marketID {
			return market
		}
	}

	// If we only have one option, pick it automatically
	if len(validMarkets) == 1 {
		return validMarkets[0]
	}

	// Print out all the markets with their index
	for index, market := range validMarkets {
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

	// Correct index as we start with 0
	index--

	if index < 0 || index > len(marketsResponse.Markets)-1 {
		fmt.Println("Invalid market selected")
		os.Exit(0)
	}

	fmt.Println("Using market:", index)

	return validMarkets[index]
}

func subscribeFeeds(dataclient api.TradingDataServiceClient) {
	// We need to subscribe to the eventbus so that we can get hold of the
	// LP for our market
	events := []eventspb.BusEventType{
		eventspb.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION,
	}

	eventBusDataReq := &api.ObserveEventBusRequest{
		Type:     events,
		MarketId: market.Id,
	}
	subscribeEventBus(dataclient, eventBusDataReq, processEventBusData)
}

type eventHandler func(api.TradingDataService_ObserveEventBusClient)

func subscribeEventBus(dataclient api.TradingDataServiceClient, eventBusDataReq *api.ObserveEventBusRequest, fn eventHandler) {
	// First we have to create the stream
	stream, err := dataclient.ObserveEventBus(context.Background())
	if err != nil {
		log.Panicln("Failed to subscribe to event bus data: ", err)
	}

	// Then we subscribe to the data
	err = stream.SendMsg(eventBusDataReq)
	if err != nil {
		log.Panicln("Unable to send event bus request on the stream", err)
	}
	go fn(stream)
}

func processEventBusData(stream api.TradingDataService_ObserveEventBusClient) {
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

		redrawRequired := false
		for _, event := range eb.Events {
			log.Println(event)
			switch event.Type {
			case eventspb.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION:
				eventLp := event.GetLiquidityProvision()
				if lp, ok := partyToLps[eventLp.Id]; ok {
					// Check if anything changed
					if lp.UpdatedAt != eventLp.UpdatedAt {
						partyToLps[eventLp.Id] = eventLp
						redrawRequired = true
					}
				} else {
					// New LP
					partyToLps[eventLp.Id] = eventLp
					redrawRequired = true
				}
			}
		}
		if redrawRequired {
			drawScreen()
		}
	}
}

// Run is the main entry point for this tool
func Run(gRPCAddress, marketID string) error {
	f, err := os.OpenFile("liquidity.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Create connection to vega
	connection, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("failed to connect to the vega gRPC port: %s", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	// Look up all the markets on this node
	market = getMarketToDisplay(dataclient, marketID)
	if market == nil {
		return fmt.Errorf("failed to get market details")
	}

	marketData = getMarketData(dataclient, market.Id)

	lps := getLiquidityProvisions(dataclient, market.Id)
	partyToLps = make(map[string]*proto.LiquidityProvision)
	for _, lp := range lps {
		partyToLps[lp.Id] = lp
	}

	initialiseScreen()

	subscribeFeeds(dataclient)

	for {
		switch ev := ts.PollEvent().(type) {
		case *tcell.EventResize:
			ts.Sync()
			drawScreen()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape ||
				ev.Rune() == 'q' {
				ts.Fini()
				os.Exit(0)
			}
		}
	}
}
