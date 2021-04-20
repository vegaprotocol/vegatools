package marketstakeviewer

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/vegaprotocol/api/grpc/clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api/grpc/clients/go/generated/code.vegaprotocol.io/vega/proto/api"

	"github.com/gdamore/tcell/v2"
	"google.golang.org/grpc"
)

var (
	ts          tcell.Screen
	redStyle    tcell.Style
	greenStyle  tcell.Style
	yellowStyle tcell.Style
	whiteStyle  tcell.Style

	args struct {
		gRPCAddress string
	}
)

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
	yellowStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorYellow)
	redStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorRed)

	return nil
}

func drawLeftString(x, y, maxw int, style tcell.Style, str string) {
	for i, c := range str {
		if i >= maxw {
			break
		}
		ts.SetContent(x+i, y, c, nil, style)
	}
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawLeftString(w-8, 0, 8, whiteStyle, text)
}

func drawRightString(x, y, w int, style tcell.Style, text string) {
	if len(text) > w {
		text = text[:w]
	}
	drawLeftString(x+w-len(text), y, w, style, text)
}

func drawRightInt64(x, y, w int, style tcell.Style, n int64) {
	f := fmt.Sprintf("%%%dd", w)
	text := fmt.Sprintf(f, n)
	drawLeftString(x, y, w, style, text)
}

func drawRightUInt64(x, y, w int, style tcell.Style, n uint64) {
	f := fmt.Sprintf("%%%dd", w)
	text := fmt.Sprintf(f, n)
	drawLeftString(x, y, w, style, text)
}

func drawRightFloat64(x, y, w int, style tcell.Style, n float64) {
	f := fmt.Sprintf("%%%d.2f", w)
	text := fmt.Sprintf(f, n)
	drawLeftString(x, y, w, style, text)
}

func drawRightPercent(x, y, w int, style tcell.Style, n float64) {
	f := fmt.Sprintf("%%%d.2f%%%%", w-1)
	text := fmt.Sprintf(f, n)
	drawLeftString(x, y, w, style, text)
}

func drawHeaders(triggerRatio float64) {
	drawLeftString(0, 0, 999, whiteStyle, fmt.Sprintf("============== Stake (trigger at %5.2f%%) =================", triggerRatio))
	drawRightString(0, 1, 15, whiteStyle, "Supplied")
	drawRightString(16, 1, 15, whiteStyle, "Target")
	drawRightString(32, 1, 15, whiteStyle, "Surplus")
	drawRightString(48, 1, 10, whiteStyle, "Percent")
	drawLeftString(59, 1, 18, whiteStyle, "Trading Mode")
	drawLeftString(78, 1, 11, whiteStyle, "Trigger")
	drawLeftString(90, 1, 64, whiteStyle, "Market ID")
}

func processMarketsStake(stream api.TradingDataService_ObserveEventBusClient, triggerRatio float64) {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed by server: ", err)
			break
		}
		if err != nil {
			log.Println("stream closed: ", err)
			break
		}

		// w, h := ts.Size()

		ts.Clear()
		drawHeaders(triggerRatio * 100.0)
		drawTime()

		for i, evt := range resp.Events {
			md := evt.GetMarketData()
			if md == nil {
				// somehow this is not a MarketData event
				continue
			}
			supplied, err := strconv.ParseUint(md.SuppliedStake, 10, 64)
			if err != nil {
				drawLeftString(0, i+2, 999, whiteStyle, err.Error())
				continue
			}
			target, err := strconv.ParseUint(md.TargetStake, 10, 64)
			if err != nil {
				drawLeftString(0, i+2, 999, whiteStyle, err.Error())
				continue
			}
			surplus := int64(supplied) - int64(target)
			drawRightUInt64(0, i+2, 15, whiteStyle, supplied)
			drawRightUInt64(16, i+2, 15, whiteStyle, target)
			var s tcell.Style
			ratio := float64(supplied) / float64(target)
			if ratio >= 1.0 {
				s = greenStyle
			} else if ratio >= triggerRatio {
				s = yellowStyle
			} else {
				s = redStyle
			}
			drawRightInt64(32, i+2, 15, s, surplus)
			drawRightPercent(48, i+2, 10, s, ratio*100.0)
			drawLeftString(59, i+2, 18, whiteStyle, fmt.Sprintf("%s", md.MarketTradingMode)[13:])
			drawLeftString(78, i+2, 11, whiteStyle, fmt.Sprintf("%s", md.Trigger)[16:])
			drawLeftString(90, i+2, 64, whiteStyle, md.Market)
		}
		ts.Show()
	}
}

func getTargetStakeTriggeringRatio(dataclient api.TradingDataServiceClient) (ratio float64, err error) {
	const key = "market.liquidity.targetstake.triggering.ratio"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var response *api.NetworkParametersResponse
	response, err = dataclient.NetworkParameters(ctx, &api.NetworkParametersRequest{})
	if err != nil {
		err = fmt.Errorf("gRPC NetworkParameters call failed: %w", err)
		return
	}
	for _, p := range response.NetworkParameters {
		if p.Key != key {
			continue
		}
		ratio, err = strconv.ParseFloat(p.Value, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse float: %w", err)
		}
		return
	}
	err = fmt.Errorf("failed to find network parameter: %s", key)
	return
}

// Run is the main entry point for this tool
func Run(gRPCAddress string) error {
	// Create connection to vega
	connection, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("Failed to connect to the vega gRPC port: %w", err)
	}
	defer connection.Close()
	dataclient := api.NewTradingDataServiceClient(connection)

	triggerRatio, err := getTargetStakeTriggeringRatio(dataclient)
	if err != nil {
		return fmt.Errorf("Failed to get target stake triggering ratio: %w", err)
	}

	observerEvent := api.ObserveEventBusRequest{
		Type: []proto.BusEventType{proto.BusEventType_BUS_EVENT_TYPE_MARKET_DATA},
	}
	eventStream, err := dataclient.ObserveEventBus(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to connect to the event bus: %w", err)
	}
	eventStream.Send(&observerEvent)
	eventStream.CloseSend()

	initialiseScreen()
	ts.Clear()
	drawHeaders(triggerRatio * 100.0)
	drawTime()
	ts.Show()

	go processMarketsStake(eventStream, triggerRatio)

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
