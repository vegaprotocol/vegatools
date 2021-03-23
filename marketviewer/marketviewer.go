package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.vegaprotocol.io/vega/proto"
	protoapi "code.vegaprotocol.io/vega/proto/api"
	"github.com/gbin/goncurses"
	"google.golang.org/grpc"
)

type commandFlags struct {
	gRPCAddress string
	marketID    string
}

var (
	scr    *goncurses.Window
	flags  *commandFlags
	market *proto.Market
)

func processFlags() {
	// Read any options from the command line
	flag.StringVar(&flags.gRPCAddress, "grpc", "n08.testnet.vega.xyz:3002", "IP:Port of vega gRPC")
	flag.StringVar(&flags.marketID, "market", "LBXRA65PN4FN5HBWRI2YBCOYDG2PBGYU", "MarketID")
	flag.Parse()
}

func getMarketToDisplay(dataclient protoapi.TradingDataServiceClient) *proto.Market {
	marketsRequest := &protoapi.MarketsRequest{}

	marketsResponse, err := dataclient.Markets(context.Background(), marketsRequest)
	if err != nil {
		return nil
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

func getMarketDepth(dataclient protoapi.TradingDataServiceClient) {
	req := &protoapi.MarketDepthSubscribeRequest{
		MarketId: flags.marketID,
	}
	stream, err := dataclient.MarketDepthSubscribe(context.Background(), req)
	if err != nil {
		log.Fatalln("Failed to subscribe to trades: ", err)
		os.Exit(0)
	}

	initialiseScreen()
	drawHeaders()
	drawTime()
	scr.Refresh()

	// Run in background and process messages
	processMarketDepth(stream)
}

func getMarketDepthUpdates(dataclient protoapi.TradingDataServiceClient) {

	req := &protoapi.MarketDepthUpdatesSubscribeRequest{
		MarketId: flags.marketID,
	}
	stream, err := dataclient.MarketDepthUpdatesSubscribe(context.Background(), req)
	if err != nil {
		log.Fatalln("Failed to subscribe to trades: ", err)
	}

	initialiseScreen()
	drawHeaders()
	drawTime()
	scr.Refresh()

	// Run in background and process messages
	processMarketDepthUpdates(stream)
}

func processMarketDepthUpdates(stream protoapi.TradingDataService_MarketDepthUpdatesSubscribeClient) {
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
	scr2, err := goncurses.Init()
	scr = scr2

	if err != nil {
		log.Fatal("init:", err)
		return err
	}

	if err := goncurses.StartColor(); err != nil {
		log.Fatal(err)
		return err
	}

	goncurses.UseDefaultColors()
	goncurses.Cursor(0)
	goncurses.Echo(false)

	goncurses.InitPair(1, goncurses.C_WHITE, -1)
	goncurses.InitPair(2, goncurses.C_RED, -1)
	goncurses.InitPair(3, goncurses.C_GREEN, -1)

	scr.AttrSet(goncurses.A_NORMAL)
	scr.SetBackground(goncurses.Char(' ') | goncurses.ColorPair(1))

	return nil
}

func drawHeaders() {
	y, x := scr.MaxYX()

	// Draw the headings
	scr.ColorOn(1)
	scr.MovePrintf(1, (x/4)-2, "Bids")
	scr.MovePrintf(1, (3*x/4)-2, "Asks")

	scr.MovePrintf(2, (x/4)-19, "--Volume--")
	scr.MovePrintf(2, (x/4)+8, "---Price---")
	scr.MovePrintf(2, (3*x/4)-22, "---Price---")
	scr.MovePrintf(2, (3*x/4)+9, "--Volume--")

	// If we have a market name, use that
	if market != nil {
		scr.MovePrintf(0, 0, "Market: %s [%s]", market.TradableInstrument.Instrument.Name, flags.marketID)
	} else {
		scr.MovePrintf(0, 0, "Market: %s", flags.marketID)
	}
	scr.MovePrintf(0, x-26, "Last Update Time:")

	scr.MovePrintf(y-1, (x/4)-8, "Volume:")
	scr.MovePrintf(y-1, (3*x/4)-8, "Volume:")
	scr.ColorOff(1)
}

func drawTime() {
	now := time.Now()
	_, x := scr.MaxYX()
	scr.MovePrintf(0, x-8, "%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
}

func drawSequenceNumber(seqNum uint64) {
	_, x := scr.MaxYX()
	scr.MovePrintf(0, (x/2)-6, "SeqNum:%6d", seqNum)
}

func processMarketDepth(stream protoapi.TradingDataService_MarketDepthSubscribeClient) {
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

		y, x := scr.MaxYX()

		scr.Clear()
		drawHeaders()
		drawTime()
		drawSequenceNumber(o.MarketDepth.SequenceNumber)

		var bidVolume uint64 = 0
		var askVolume uint64 = 0

		// Print Buys
		scr.ColorOn(3)
		buyPriceLevels := o.MarketDepth.Buy
		for index, pl := range buyPriceLevels {
			bidVolume += pl.Volume
			if index > (y - 6) {
				continue
			}
			scr.MovePrintf(index+3, (x/4)-21, "%12d", pl.Volume)
			scr.MovePrintf(index+3, (x/4)+7, "%12d", pl.Price)
		}
		scr.ColorOff(3)

		// Print Sells
		scr.ColorOn(2)
		sellPriceLevels := o.MarketDepth.Sell
		for index, pl := range sellPriceLevels {
			askVolume += pl.Volume
			if index > (y - 6) {
				continue
			}
			scr.MovePrintf(index+3, (3*x/4)-22, "%d", pl.Price)
			scr.MovePrintf(index+3, (3*x/4)+9, "%d", pl.Volume)
		}
		scr.ColorOff(2)

		scr.MovePrintf(y-1, (x / 4), "%8d", bidVolume)
		scr.MovePrintf(y-1, (3 * x / 4), "%8d", askVolume)

		scr.Refresh()
	}
}

func main() {
	// Read any options from the command line
	flags = &commandFlags{}
	processFlags()

	// Create connection to vega
	connection, err := grpc.Dial(flags.gRPCAddress, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		log.Println("Failed to connect to the vega gRPC port: ", err)
		os.Exit(1)
	}
	defer connection.Close()
	dataclient := protoapi.NewTradingDataServiceClient(connection)

	// Look up all the markets on this node
	m := getMarketToDisplay(dataclient)
	if m != nil {
		market = m
		flags.marketID = m.Id
	}

	// Install a Ctrl-C handler to tidy up the screen
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		goncurses.End()
		os.Exit(0)
	}()

	// Start the book displaying
	getMarketDepth(dataclient)
}
