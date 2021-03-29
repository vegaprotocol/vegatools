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
	//party          string
	mapMarketToLPs map[string][]*proto.LiquidityProvision = map[string][]*proto.LiquidityProvision{}
	lp             *proto.LiquidityProvision
	auction        *proto.AuctionEvent
	acctMargin     uint64
	acctGeneral    uint64
	acctBond       uint64
	mapOrders      map[string]*proto.Order = map[string]*proto.Order{}
	marketData     *proto.MarketData
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

	return marketsResponse.Markets[index]
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
		log.Println(err)
		return
	}

	for _, acct := range response.Accounts {
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

func drawScreen() {
	ts.Clear()
	drawHeaders()
	drawTime()
	drawLP()
	drawAccounts()
	drawOrders()
	drawMarketState()
	ts.Show()
}

func drawHeaders() {
	w, _ := ts.Size()

	// If we have a market name, use that
	if market != nil {
		text := fmt.Sprintf("Market: %s [%s]", market.TradableInstrument.Instrument.Name, market.Id)
		drawString(0, 0, whiteStyle, text)
	} else {
		text := fmt.Sprintf("Market: %s", market.Id)
		drawString(0, 0, whiteStyle, text)
	}
	drawString(w-26, 0, whiteStyle, "Last Update Time:")
}

func drawLP() {
	w, _ := ts.Size()
	hw := w / 2
	buyStartRow := 2
	buyTitle := "Buy Side Shape"
	drawString((w/4)-(len(buyTitle)/2), buyStartRow, greenStyle, buyTitle)
	drawString(0, buyStartRow+1, whiteStyle, "OrderID")
	drawString(hw/4, buyStartRow+1, whiteStyle, "Reference")
	drawString(hw/2, buyStartRow+1, whiteStyle, "Offset")
	drawString((3*hw)/4, buyStartRow+1, whiteStyle, "Proportion")
	for index, lor := range lp.Buys {
		buyRow := buyStartRow + index + 2
		drawString(0, buyRow, whiteStyle, lor.OrderId)
		drawString(hw/4, buyRow, whiteStyle, lor.LiquidityOrder.Reference.String())
		offset := strconv.Itoa(int(lor.LiquidityOrder.Offset))
		drawString(hw/2, buyRow, whiteStyle, offset)
		proportion := strconv.Itoa(int(lor.LiquidityOrder.Proportion))
		drawString((3*hw)/4, buyRow, whiteStyle, proportion)
	}

	sellStartRow := 2
	sellTitle := "Sell Side Shape"
	drawString((3*w)/4-(len(sellTitle)/2), sellStartRow, redStyle, sellTitle)
	drawString(hw, buyStartRow+1, whiteStyle, "OrderID")
	drawString(hw+(hw/4), buyStartRow+1, whiteStyle, "Reference")
	drawString(hw+(hw/2), buyStartRow+1, whiteStyle, "Offset")
	drawString(hw+((3*hw)/4), buyStartRow+1, whiteStyle, "Proportion")
	for index, lor := range lp.Sells {
		sellRow := sellStartRow + index + 2
		drawString(hw, sellRow, whiteStyle, lor.OrderId)
		drawString(hw+(hw/4), sellRow, whiteStyle, lor.LiquidityOrder.Reference.String())
		offset := strconv.Itoa(int(lor.LiquidityOrder.Offset))
		drawString(hw+(hw/2), sellRow, whiteStyle, offset)
		proportion := strconv.Itoa(int(lor.LiquidityOrder.Proportion))
		drawString(hw+((hw*3)/4), sellRow, whiteStyle, proportion)
	}
}

// Bottom row of display
func drawMarketState() {
	if marketData == nil {
		return
	}

	w, h := ts.Size()

	text := fmt.Sprintf("Market State: %s", marketData.MarketTradingMode.String())
	drawString((w+len(text))/2, 0, whiteStyle, text)

	drawString(w-1, h-1, redStyle, "*")
}

func drawAccounts() {
	w, h := ts.Size()

	text := fmt.Sprintf("General Account %d", acctGeneral)
	drawString((0*w)/3, h-1, whiteStyle, text)
	text = fmt.Sprintf("Margin Account %d", acctMargin)
	drawString((w-len(text))/2, h-1, whiteStyle, text)
	text = fmt.Sprintf("Bond Account %d", acctBond)
	drawString(w-len(text)-1, h-1, whiteStyle, text)
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}

func drawOrders() {
	w, h := ts.Size()
	hw := w / 2
	buyStartRow := h / 2

	// Buy header
	drawString(0, buyStartRow, whiteStyle, "OrderID")
	drawString(hw/4, buyStartRow, whiteStyle, "Price")
	drawString(hw/2, buyStartRow, whiteStyle, "Size")
	drawString((3*hw)/4, buyStartRow, whiteStyle, "Remaining")

	sellStartRow := h / 2
	drawString(hw, sellStartRow, whiteStyle, "OrderID")
	drawString(hw+(hw/4), sellStartRow, whiteStyle, "Price")
	drawString(hw+(hw/2), sellStartRow, whiteStyle, "Size")
	drawString(hw+((3*hw)/4), sellStartRow, whiteStyle, "Remaining")

	for _, order := range mapOrders {
		if order != nil {
			if order.Side == proto.Side_SIDE_BUY {
				buyStartRow++
				drawString(0, buyStartRow, whiteStyle, order.Id)
				drawString(hw/4, buyStartRow, whiteStyle, strconv.FormatUint(order.Price, 10))
				drawString(hw/2, buyStartRow, whiteStyle, strconv.FormatUint(order.Size, 10))
				drawString((3*hw)/4, buyStartRow, whiteStyle, strconv.FormatUint(order.Remaining, 10))
			} else {
				sellStartRow++
				drawString(hw, sellStartRow, whiteStyle, order.Id)
				drawString(hw+(hw/4), sellStartRow, whiteStyle, strconv.FormatUint(order.Price, 10))
				drawString(hw+(hw/2), sellStartRow, whiteStyle, strconv.FormatUint(order.Size, 10))
				drawString(hw+((3*hw)/4), sellStartRow, whiteStyle, strconv.FormatUint(order.Remaining, 10))
			}
		}
	}
}

func subscribeFeeds(dataclient api.TradingDataServiceClient) {
	// We need to subscribe to the eventbus so that we can get hold of the
	// LP, order and margin updates for a party
	events := []proto.BusEventType{proto.BusEventType_BUS_EVENT_TYPE_MARKET_DATA,
		proto.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION,
		proto.BusEventType_BUS_EVENT_TYPE_ORDER,
		proto.BusEventType_BUS_EVENT_TYPE_ACCOUNT}
	subscribeEventBus(dataclient, events)
}

func subscribeEventBus(dataclient api.TradingDataServiceClient, events []proto.BusEventType) {
	eventBusDataReq := &api.ObserveEventBusRequest{
		Type: events,
		//		MarketId: lp.MarketId,
		PartyId: lp.PartyId,
	}
	// First we have to create the stream
	stream, err := dataclient.ObserveEventBus(context.Background())
	if err != nil {
		log.Fatalln("Failed to subscribe to event bus data: ", err)
	}

	// Then we subscribe to the data
	err = stream.SendMsg(eventBusDataReq)
	if err != nil {
		log.Fatalln("Unable to send event bus request on the stream", err)
	}
	go processEventBusData(stream)
}

func processEventBusData(stream api.TradingDataService_ObserveEventBusClient) {
	for {
		eb, err := stream.Recv()
		if err == io.EOF {
			log.Println("event bus data: stream closed by server err:", err)
			break
		}
		if err != nil {
			log.Println("event bus data: stream closed err:", err)
			break
		}

		for _, event := range eb.Events {
			switch event.Type {
			case proto.BusEventType_BUS_EVENT_TYPE_MARKET_DATA:
				marketData = event.GetMarketData()
				drawScreen()
			case proto.BusEventType_BUS_EVENT_TYPE_LIQUIDITY_PROVISION:
				lp = event.GetLiquidityProvision()
				drawScreen()
			case proto.BusEventType_BUS_EVENT_TYPE_ORDER:
				// Check we are interested in this order
				if _, ok := mapOrders[event.GetOrder().Id]; ok {
					mapOrders[event.GetOrder().Id] = event.GetOrder()
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

func populateOrderMap() {
	// Go through the lp and extract the order IDs
	for _, lo := range lp.Buys {
		mapOrders[lo.OrderId] = nil
	}

	for _, lo := range lp.Sells {
		mapOrders[lo.OrderId] = nil
	}
}

// Run is the main entry point for this tool
func Run(gRPCAddress, marketID, partyID string) error {
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

	getAccountDetails(dataclient, market.Id, partyID, market.TradableInstrument.Instrument.GetFuture().GetSettlementAsset())

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
