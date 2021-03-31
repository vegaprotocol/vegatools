package liquidityviewer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

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
	//party          string
	mapMarketToLPs map[string][]*proto.LiquidityProvision = map[string][]*proto.LiquidityProvision{}
	lp             *proto.LiquidityProvision
	auction        *proto.AuctionEvent
	acctMargin     uint64
	acctGeneral    uint64
	acctBond       uint64
	mapOrders      map[string]*proto.Order = map[string]*proto.Order{}
	marketData     *proto.MarketData
	position       *proto.Position
)

func getLiquidityProvisions(dataclient api.TradingDataServiceClient, marketId string) []*proto.LiquidityProvision {
	lpReq := &api.LiquidityProvisionsRequest{Market: marketId}

	response, err := dataclient.LiquidityProvisions(context.Background(), lpReq)
	if err != nil {
		log.Println(err)
	}
	return response.LiquidityProvisions
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

	return validMarkets[index]
}

func getPartyToDisplay(dataclient api.TradingDataServiceClient, marketID, partyID string) *proto.LiquidityProvision {
	// We cached the LPs earlier so we can extract the parties from there
	lps := mapMarketToLPs[marketID]

	// If the user has picked a party already that is valid, use that
	for _, lp := range lps {
		if lp.PartyId == partyID {
			return lp
		}
	}

	// If we only have one option, choose it automatically
	if len(lps) == 1 {
		return lps[0]
	}

	// Print out all the parties with their index
	for index, lp := range lps {
		fmt.Printf("[%d]:%s\n", index, lp.PartyId)
	}

	// Ask the user to select a party
	fmt.Printf("Which party do you want to view? ")

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

	if index < 0 || index > len(lps)-1 {
		fmt.Println("Invalid party selected")
		return nil
	}
	return lps[index]
}

func getAccountDetails(dataclient api.TradingDataServiceClient, marketId, partyId, assetId string) {
	lpReq := &api.PartyAccountsRequest{PartyId: partyId,
		Asset: assetId}

	response, err := dataclient.PartyAccounts(context.Background(), lpReq)
	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, acct := range response.Accounts {
		log.Println(acct)
		switch acct.Type {
		case proto.AccountType_ACCOUNT_TYPE_BOND:
			acctBond = acct.Balance
		case proto.AccountType_ACCOUNT_TYPE_MARGIN:
			acctMargin = acct.Balance
		case proto.AccountType_ACCOUNT_TYPE_GENERAL:
			acctGeneral = acct.Balance
		}
	}
}

func subscribePositions(dataclient api.TradingDataServiceClient, marketID string, userKey string) {
	req := &api.PositionsSubscribeRequest{
		MarketId: marketID,
		PartyId:  userKey,
	}
	stream, err := dataclient.PositionsSubscribe(context.Background(), req)
	if err != nil {
		log.Panicln("Failed to subscribe to positions: ", err)
	}

	// Run in background and process messages
	go processPositions(stream)
}

func processPositions(stream api.TradingDataService_PositionsSubscribeClient) {
	for {
		o, err := stream.Recv()
		if err == io.EOF {
			log.Panicln("positions: stream closed by server err:", err)
			break
		}
		if err != nil {
			log.Panicln("positions: stream closed err:", err)
			break
		}
		position = o.GetPosition()
	}
}

func subscribeFeeds(dataclient api.TradingDataServiceClient) {
	// We need to subscribe to the eventbus so that we can get hold of the
	// LP, order and margin updates for a party
	events := []proto.BusEventType{proto.BusEventType_BUS_EVENT_TYPE_MARKET_DATA}

	eventBusDataReq := &api.ObserveEventBusRequest{
		Type:     events,
		MarketId: market.Id,
	}
	subscribeEventBus(dataclient, eventBusDataReq, processEventBusData)

	events2 := []proto.BusEventType{proto.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION,
		proto.BusEventType_BUS_EVENT_TYPE_ORDER,
		proto.BusEventType_BUS_EVENT_TYPE_ACCOUNT}

	eventBusDataReq2 := &api.ObserveEventBusRequest{
		Type:    events2,
		PartyId: lp.PartyId,
	}
	subscribeEventBus(dataclient, eventBusDataReq2, processEventBusData2)

	subscribePositions(dataclient, market.Id, lp.PartyId)
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

		for _, event := range eb.Events {
			log.Println(event)
			switch event.Type {
			case proto.BusEventType_BUS_EVENT_TYPE_MARKET_DATA:
				marketData = event.GetMarketData()
				drawScreen()
			}
		}
	}
}

func processEventBusData2(stream api.TradingDataService_ObserveEventBusClient) {
	for {
		eb, err := stream.Recv()
		if err == io.EOF {
			log.Panicln("event bus data2: stream closed by server err:", err)
			break
		}
		if err != nil {
			log.Panicln("event bus data2: stream closed err:", err)
			break
		}

		for _, event := range eb.Events {
			log.Println(event)
			switch event.Type {
			case proto.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION:
				lp = event.GetLiquidityProvision()
				populateOrderMap()
				drawScreen()
			case proto.BusEventType_BUS_EVENT_TYPE_ORDER:
				// Check we are interested in this order
				if processOrder(event.GetOrder()) {
					drawScreen()
				}
			case proto.BusEventType_BUS_EVENT_TYPE_ACCOUNT:
				account := event.GetAccount()
				switch account.Type {
				case proto.AccountType_ACCOUNT_TYPE_BOND:
					acctBond = account.Balance
				case proto.AccountType_ACCOUNT_TYPE_GENERAL:
					acctGeneral = account.Balance
				case proto.AccountType_ACCOUNT_TYPE_MARGIN:
					acctMargin = account.Balance
				}
				drawScreen()
			}
		}
	}
}

func processOrder(order *proto.Order) bool {
	if _, ok := mapOrders[order.Id]; ok {
		if order.Status != proto.Order_STATUS_ACTIVE ||
			order.Remaining == 0 {
			delete(mapOrders, order.Id)
		} else {
			mapOrders[order.Id] = order
		}
	}
	return true
}

func populateOrderMap() {
	// Go through the lp and extract the order IDs
	for _, lo := range lp.Buys {
		if _, ok := mapOrders[lo.OrderId]; !ok {
			mapOrders[lo.OrderId] = nil
		}
	}

	for _, lo := range lp.Sells {
		if _, ok := mapOrders[lo.OrderId]; !ok {
			mapOrders[lo.OrderId] = nil
		}
	}
}

// Run is the main entry point for this tool
func Run(gRPCAddress, marketID, partyID string) error {
	f, err := os.OpenFile("liquidity.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	// Look up all the parties in this market
	lp = getPartyToDisplay(dataclient, market.Id, partyID)
	if lp == nil {
		return fmt.Errorf("failed to get the party details")
	}

	getAccountDetails(dataclient, market.Id, lp.PartyId, market.TradableInstrument.Instrument.GetFuture().GetSettlementAsset())

	populateOrderMap()

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
