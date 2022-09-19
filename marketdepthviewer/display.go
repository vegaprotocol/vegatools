package marketdepthviewer

import (
	"fmt"
	"log"
	"math/big"
	"sort"
	"time"

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
	inverseRedStyle = tcell.StyleDefault.
		Background(tcell.ColorRed).
		Foreground(tcell.ColorBlack)

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
	drawString((w/4)-2, 2, whiteStyle, "Bids")
	drawString((3*w/4)-2, 2, whiteStyle, "Asks")

	drawString((w/4)-19, 3, whiteStyle, "--Volume--")
	drawString((w/4)+8, 3, whiteStyle, "---Price---")
	drawString((3*w/4)-22, 3, whiteStyle, "---Price---")
	drawString((3*w/4)+9, 3, whiteStyle, "--Volume--")

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

	drawString((w/4)-8, h-1, whiteStyle, "Volume:")
	drawString((3*w/4)-8, h-1, whiteStyle, "Volume:")
}

func drawMarketState() {
	if marketData != nil {
		w, h := ts.Size()
		text := marketData.MarketTradingMode.String()
		drawString((w-len(text))/2, h-1, whiteStyle, text)
		text = fmt.Sprintf("Open Interest: %d", marketData.OpenInterest)
		drawString(w-len(text), 1, whiteStyle, text)
		ts.Show()
	}
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}

func drawSequenceNumber(seqNum uint64) {
	w, _ := ts.Size()
	text := fmt.Sprintf("%s SeqNum:%6d", updateMode, seqNum)
	drawString((w/2)-6, 0, whiteStyle, text)
}

func drawMarketDepth() {
	displayMutex.Lock()
	defer displayMutex.Unlock()

	// Get the keys and sort
	var buyPrices []*big.Int = make([]*big.Int, 0, len(book.buys))
	var sellPrices []*big.Int = make([]*big.Int, 0, len(book.sells))

	for p := range book.buys {
		price, _ := new(big.Int).SetString(p, 10)
		buyPrices = append(buyPrices, price)
	}
	if len(buyPrices) > 1 {
		sort.SliceStable(buyPrices, func(i, j int) bool {
			return buyPrices[i].Cmp(buyPrices[j]) == 1
		})
	}

	for p := range book.sells {
		price, _ := new(big.Int).SetString(p, 10)
		sellPrices = append(sellPrices, price)
	}
	if len(sellPrices) > 1 {
		sort.SliceStable(sellPrices, func(i, j int) bool {
			return sellPrices[i].Cmp(sellPrices[j]) == -1
		})
	}

	w, h := ts.Size()

	ts.Clear()
	drawHeaders()
	drawTime()
	drawSequenceNumber(book.seqNum)
	drawMarketState()

	var bidVolume uint64
	var askVolume uint64

	// Print Buys
	for index, price := range buyPrices {
		pl := book.buys[price.String()]
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
	for index, price := range sellPrices {
		pl := book.sells[price.String()]
		askVolume += pl.Volume
		if index > (h - 6) {
			continue
		}
		drawString((3*w/4)-22, index+4, redStyle, pl.Price)
		text := fmt.Sprintf("%d", pl.Volume)
		drawString((3*w/4)+9, index+4, redStyle, text)
	}

	text := fmt.Sprintf("%8d", bidVolume)
	drawString((w / 4), h-1, whiteStyle, text)
	text = fmt.Sprintf("%8d", askVolume)
	drawString((3 * w / 4), h-1, whiteStyle, text)

	ts.Show()
	dirty = false
	lastRedraw = time.Now()
}
