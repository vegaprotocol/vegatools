package liquidityviewer

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vegaprotocol/api/go/generated/code.vegaprotocol.io/vega/proto"
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
		text := fmt.Sprintf("Market: %s", market.TradableInstrument.Instrument.Name)
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
	hw := w / 2
	buyStartRow := 3
	buyTitle := "Buy Side Shape"
	drawString((w/4)-(len(buyTitle)/2), buyStartRow, greenStyle, buyTitle)
	drawString(0, buyStartRow+1, whiteStyle, "OrderID")
	drawString(hw/4, buyStartRow+1, whiteStyle, "Reference")
	drawString(hw/2, buyStartRow+1, whiteStyle, "Offset")
	drawString((3*hw)/4, buyStartRow+1, whiteStyle, "Proportion")
	for index, lor := range lp.Buys {
		buyRow := buyStartRow + index + 2
		drawString(0, buyRow, whiteStyle, lor.OrderId)
		drawString(hw/4, buyRow, whiteStyle, getReferenceStr(lor.LiquidityOrder.Reference))
		offset := strconv.Itoa(int(lor.LiquidityOrder.Offset))
		drawString(hw/2, buyRow, whiteStyle, offset)
		proportion := strconv.Itoa(int(lor.LiquidityOrder.Proportion))
		drawString((3*hw)/4, buyRow, whiteStyle, proportion)
	}

	sellStartRow := 3
	sellTitle := "Sell Side Shape"
	drawString((3*w)/4-(len(sellTitle)/2), sellStartRow, redStyle, sellTitle)
	drawString(hw, buyStartRow+1, whiteStyle, "OrderID")
	drawString(hw+(hw/4), buyStartRow+1, whiteStyle, "Reference")
	drawString(hw+(hw/2), buyStartRow+1, whiteStyle, "Offset")
	drawString(hw+((3*hw)/4), buyStartRow+1, whiteStyle, "Proportion")
	for index, lor := range lp.Sells {
		sellRow := sellStartRow + index + 2
		drawString(hw, sellRow, whiteStyle, lor.OrderId)
		drawString(hw+(hw/4), sellRow, whiteStyle, getReferenceStr(lor.LiquidityOrder.Reference))
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

	w, _ := ts.Size()

	text := fmt.Sprintf("Market State: %s", marketData.MarketTradingMode.String())
	drawString((w-len(text))/3, 0, whiteStyle, text)

	text = fmt.Sprintf("Commitment: %d", lp.CommitmentAmount)
	drawString(((w-len(text))*2)/3, 0, whiteStyle, text)

	text = fmt.Sprintf("Stake (Target:%s/Suppled:%s)", marketData.TargetStake, marketData.SuppliedStake)
	drawString(w-len(text), 1, whiteStyle, text)
}

func drawAccounts() {
	w, h := ts.Size()

	text := fmt.Sprintf("General Account %d", acctGeneral)
	drawString((0*w)/3, h-1, whiteStyle, text)
	text = fmt.Sprintf("Margin Account %d", acctMargin)
	drawString((w-len(text))/2, h-1, whiteStyle, text)
	text = fmt.Sprintf("Bond Account %d", acctBond)
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
	text = fmt.Sprintf("Realised PnL %d", position.RealisedPnl)
	if position.RealisedPnl >= 0 {
		drawString((w-len(text))/2, h-2, greenStyle, text)
	} else {
		drawString((w-len(text))/2, h-2, redStyle, text)
	}

	text = fmt.Sprintf("Unrealised PnL %d", position.UnrealisedPnl)
	if position.UnrealisedPnl >= 0 {
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

	for _, order := range orders {
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
