package liquidityviewer

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	proto "code.vegaprotocol.io/protos/vega"
	"github.com/gdamore/tcell/v2"
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

func drawScreen() {
	ts.Clear()
	drawHeaders()
	drawTime()
	drawLP()
	drawAccounts()
	drawOrders()
	drawMarketState()
	drawPosition()
	ts.Show()
}

func drawHeaders() {
	w, _ := ts.Size()

	// If we have a market name, use that
	if market != nil {
		text := fmt.Sprintf("%s", market.TradableInstrument.Instrument.Name)
		drawString(0, 0, whiteStyle, text)
		drawString(0, 1, whiteStyle, market.Id)

	} else {
		text := fmt.Sprintf("Market: %s", market.Id)
		drawString(0, 0, whiteStyle, text)
	}
	drawString(w-26, 0, whiteStyle, "Last Update Time:")
}

func getReferenceStr(ref proto.PeggedReference) string {
	switch ref {
	case proto.PeggedReference_PEGGED_REFERENCE_MID:
		return "MID"
	case proto.PeggedReference_PEGGED_REFERENCE_BEST_ASK:
		return "BEST ASK"
	case proto.PeggedReference_PEGGED_REFERENCE_BEST_BID:
		return "BEST BID"
	default:
		return "N/A"
	}
}

func drawLP() {
	w, _ := ts.Size()
	buyStartRow := 3
	buyTitle := "Buy Side Shape"
	drawString((w/4)-(len(buyTitle)/2), buyStartRow, greenStyle, buyTitle)
	drawStringPc(0, buyStartRow+1, whiteStyle, "OrderID")
	drawStringPc(25, buyStartRow+1, whiteStyle, "Reference")
	drawStringPc(35, buyStartRow+1, whiteStyle, "Offset")
	drawStringPc(43, buyStartRow+1, whiteStyle, "Prop")
	for index, lor := range lp.Buys {
		buyRow := buyStartRow + index + 2
		drawStringPc(0, buyRow, whiteStyle, lor.OrderId)
		drawStringPc(25, buyRow, whiteStyle, getReferenceStr(lor.LiquidityOrder.Reference))
		offset := strconv.Itoa(int(lor.LiquidityOrder.Offset))
		drawStringPc(35, buyRow, whiteStyle, offset)
		proportion := strconv.Itoa(int(lor.LiquidityOrder.Proportion))
		drawStringPc(43, buyRow, whiteStyle, proportion)
	}

	sellStartRow := 3
	sellTitle := "Sell Side Shape"
	drawString((3*w)/4-(len(sellTitle)/2), sellStartRow, redStyle, sellTitle)
	drawStringPc(50, buyStartRow+1, whiteStyle, "OrderID")
	drawStringPc(75, buyStartRow+1, whiteStyle, "Reference")
	drawStringPc(85, buyStartRow+1, whiteStyle, "Offset")
	drawStringPc(93, buyStartRow+1, whiteStyle, "Prop")
	for index, lor := range lp.Sells {
		sellRow := sellStartRow + index + 2
		drawStringPc(50, sellRow, whiteStyle, lor.OrderId)
		drawStringPc(75, sellRow, whiteStyle, getReferenceStr(lor.LiquidityOrder.Reference))
		offset := strconv.Itoa(int(lor.LiquidityOrder.Offset))
		drawStringPc(85, sellRow, whiteStyle, offset)
		proportion := strconv.Itoa(int(lor.LiquidityOrder.Proportion))
		drawStringPc(93, sellRow, whiteStyle, proportion)
	}
}

// Bottom row of display
func drawMarketState() {
	if marketData == nil {
		return
	}

	w, _ := ts.Size()

	text := fmt.Sprintf("%s", marketData.MarketTradingMode.String())
	drawString((w-len(text))/3, 0, whiteStyle, text)

	text = fmt.Sprintf("Commitment: %s", lp.CommitmentAmount)
	drawString(((w-len(text))*2)/3, 0, whiteStyle, text)

	text = fmt.Sprintf("Target Stake:%s", marketData.TargetStake)
	drawString(w-len(text), 1, whiteStyle, text)

	text = fmt.Sprintf("Supplied Stake:%s", marketData.SuppliedStake)
	drawString(w-len(text), 2, whiteStyle, text)
}

func drawAccounts() {
	w, h := ts.Size()

	text := fmt.Sprintf("General Account %s", acctGeneral)
	drawString((0*w)/3, h-1, whiteStyle, text)
	text = fmt.Sprintf("Margin Account %s", acctMargin)
	drawString((w-len(text))/2, h-1, whiteStyle, text)
	text = fmt.Sprintf("Bond Account %s", acctBond)
	drawString(w-len(text), h-1, whiteStyle, text)
}

func drawPosition() {
	if position == nil {
		return
	}

	w, h := ts.Size()
	text := fmt.Sprintf("Open Volume %d", position.OpenVolume)
	if position.OpenVolume >= 0 {
		drawString(0, h-2, greenStyle, text)
	} else {
		drawString(0, h-2, redStyle, text)
	}
	text = fmt.Sprintf("Realised PnL %s", position.RealisedPnl)
	if len(position.RealisedPnl) > 0 && position.RealisedPnl[0] == '+' {
		drawString((w-len(text))/2, h-2, greenStyle, text)
	} else {
		drawString((w-len(text))/2, h-2, redStyle, text)
	}

	text = fmt.Sprintf("Unrealised PnL %s", position.UnrealisedPnl)
	if len(position.UnrealisedPnl) > 0 && position.UnrealisedPnl[0] == '+' {
		drawString(w-len(text), h-2, greenStyle, text)
	} else {
		drawString(w-len(text), h-2, redStyle, text)
	}
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}

func drawOrders() {
	_, h := ts.Size()
	startRow := h / 2

	// Buy header
	drawStringPc(0, startRow, whiteStyle, "OrderID")
	drawStringPc(25, startRow, whiteStyle, "Price")
	drawStringPc(34, startRow, whiteStyle, "Size")
	drawStringPc(42, startRow, whiteStyle, "Remain")

	drawStringPc(50, startRow, whiteStyle, "OrderID")
	drawStringPc(75, startRow, whiteStyle, "Price")
	drawStringPc(84, startRow, whiteStyle, "Size")
	drawStringPc(92, startRow, whiteStyle, "Remain")

	// Convert map into slice
	orders := []*proto.Order{}
	for _, order := range mapOrders {
		if order != nil {
			orders = append(orders, order)
		}
	}

	// Sort them
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].Id < orders[j].Id
	})

	buyStartRow := startRow
	sellStartRow := startRow

	for _, order := range orders {
		if order != nil {
			if order.Side == proto.Side_SIDE_BUY {
				buyStartRow++
				drawStringPc(0, buyStartRow, whiteStyle, order.Id)
				drawStringPc(25, buyStartRow, whiteStyle, order.Price)
				drawStringPc(34, buyStartRow, whiteStyle, strconv.FormatUint(order.Size, 10))
				drawStringPc(42, buyStartRow, whiteStyle, strconv.FormatUint(order.Remaining, 10))
			} else {
				sellStartRow++
				drawStringPc(50, sellStartRow, whiteStyle, order.Id)
				drawStringPc(75, sellStartRow, whiteStyle, order.Price)
				drawStringPc(84, sellStartRow, whiteStyle, strconv.FormatUint(order.Size, 10))
				drawStringPc(92, sellStartRow, whiteStyle, strconv.FormatUint(order.Remaining, 10))
			}
		}
	}
}
