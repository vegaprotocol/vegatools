package marketdepthviewer

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

func (m *mdv) initialiseScreen() error {
	var err error
	m.ts, err = tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create new tcell screen: %w", err)
	}

	err = m.ts.Init()
	if err != nil {
		return fmt.Errorf("failed to initialise the tcell screen: %w", err)
	}

	m.whiteStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorWhite)
	m.greenStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorGreen)
	m.redStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorRed)

	return nil
}

func (m *mdv) drawString(x, y int, style tcell.Style, str string) {
	for i, c := range str {
		m.ts.SetContent(x+i, y, c, nil, style)
	}
}

func (m *mdv) drawHeaders() {
	w, h := m.ts.Size()

	// Draw the headings
	m.drawString((w/4)-2, 2, m.whiteStyle, "Bids")
	m.drawString((3*w/4)-2, 2, m.whiteStyle, "Asks")

	m.drawString((w/4)-19, 3, m.whiteStyle, "--Volume--")
	m.drawString((w/4)+8, 3, m.whiteStyle, "---Price---")
	m.drawString((3*w/4)-22, 3, m.whiteStyle, "---Price---")
	m.drawString((3*w/4)+9, 3, m.whiteStyle, "--Volume--")

	// If we have a market name, use that
	if m.market != nil {
		text := fmt.Sprintf("Market: %s", m.market.TradableInstrument.Instrument.Name)
		m.drawString(0, 0, m.whiteStyle, text)
		m.drawString(0, 1, m.whiteStyle, m.market.Id)
	} else {
		text := fmt.Sprintf("Market: %s", m.market.Id)
		m.drawString(0, 0, m.whiteStyle, text)
	}
	m.drawString(w-26, 0, m.whiteStyle, "Last Update Time:")

	m.drawString((w/4)-8, h-1, m.whiteStyle, "Volume:")
	m.drawString((3*w/4)-8, h-1, m.whiteStyle, "Volume:")
}

func (m *mdv) drawMarketState() {
	if m.marketData != nil {
		w, h := m.ts.Size()
		text := m.marketData.MarketTradingMode.String()
		m.drawString((w-len(text))/2, h-1, m.whiteStyle, text)
		text = fmt.Sprintf("Open Interest: %d", m.marketData.OpenInterest)
		m.drawString(w-len(text), 1, m.whiteStyle, text)
		m.ts.Show()
	}
}

func (m *mdv) drawTime() {
	now := time.Now()
	w, _ := m.ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	m.drawString(w-8, 0, m.whiteStyle, text)
}

func (m *mdv) drawSequenceNumber(seqNum uint64) {
	w, _ := m.ts.Size()
	text := fmt.Sprintf("%s SeqNum:%6d", m.updateMode, seqNum)
	m.drawString((w/2)-(len(text)/2), 0, m.whiteStyle, text)
}

func (m *mdv) drawMarketDepth() {
	m.displayMutex.Lock()
	defer m.displayMutex.Unlock()

	// Get the keys and sort
	var buyPrices []*big.Int = make([]*big.Int, 0, len(m.book.buys))
	var sellPrices []*big.Int = make([]*big.Int, 0, len(m.book.sells))

	for p := range m.book.buys {
		price, _ := new(big.Int).SetString(p, 10)
		buyPrices = append(buyPrices, price)
	}
	if len(buyPrices) > 1 {
		sort.SliceStable(buyPrices, func(i, j int) bool {
			return buyPrices[i].Cmp(buyPrices[j]) == 1
		})
	}

	for p := range m.book.sells {
		price, _ := new(big.Int).SetString(p, 10)
		sellPrices = append(sellPrices, price)
	}
	if len(sellPrices) > 1 {
		sort.SliceStable(sellPrices, func(i, j int) bool {
			return sellPrices[i].Cmp(sellPrices[j]) == -1
		})
	}

	w, h := m.ts.Size()

	m.ts.Clear()
	m.drawHeaders()
	m.drawTime()
	m.drawSequenceNumber(m.book.seqNum)
	m.drawMarketState()

	var bidVolume uint64
	var askVolume uint64

	// Print Buys
	for index, price := range buyPrices {
		pl := m.book.buys[price.String()]
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
	for index, price := range sellPrices {
		pl := m.book.sells[price.String()]
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
	m.dirty = false
	m.lastRedraw = time.Now()
}
