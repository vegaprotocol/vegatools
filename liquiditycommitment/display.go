package liquiditycommitment

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"code.vegaprotocol.io/vega/protos/vega"
	"github.com/gdamore/tcell/v2"
)

func tsToDate(ts int64) string {
	timeT := time.Unix(0, ts)
	return timeT.UTC().Format("2006/01/02 03:04")
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
	greyStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorLightGrey)

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
	ts.Show()
}

func drawHeaders() {
	w, h := ts.Size()

	// If we have an instrument name, use that
	if market != nil {
		drawString(0, 0, whiteStyle, market.TradableInstrument.Instrument.Name)
	}

	text := fmt.Sprintf("Market: %s", market.Id)
	drawString((w-len(text))/2, h-1, whiteStyle, text)

	text = marketData.MarketTradingMode.String()
	drawString((w-len(text))/2, 0, whiteStyle, text)

	// Last update time
	drawString(w-26, 0, whiteStyle, "Last Update Time:")

	// Stake values
	ts := fmt.Sprintf("Target Commitment: %s", marketData.TargetStake)
	ss := fmt.Sprintf("Supplied Commitment: %s", marketData.SuppliedStake)
	drawString(0, 1, whiteStyle, ts)
	drawString(w-len(ss), 1, whiteStyle, ss)

	// LP header
	drawString(0, 3, whiteStyle, "PartyID")
	drawStringPc(45, 3, whiteStyle, "Status")
	drawStringPc(55, 3, whiteStyle, "Created")
	drawStringPc(70, 3, whiteStyle, "Modified")
	drawStringPc(85, 3, whiteStyle, "Fee")
	drawString(w-10, 3, whiteStyle, "Commitment")
}

func drawLP() {
	w, h := ts.Size()

	// Sort the LPs by fee
	var lps []*vega.LiquidityProvision
	for _, lp := range partyToLps {
		lps = append(lps, lp)
	}

	sort.Slice(lps, func(i, j int) bool {
		return lps[i].PartyId < lps[j].PartyId
	})

	row := 4
	for count, lp := range lps {
		style := greyStyle
		if row%2 == 0 {
			style = whiteStyle
		}

		status := lp.Status.String()
		status = strings.Replace(status, "STATUS_", "", 1)

		drawString(0, row, style, lp.PartyId)
		drawStringPc(45, row, style, status)
		drawStringPc(55, row, style, tsToDate(lp.CreatedAt))
		drawStringPc(70, row, style, tsToDate(lp.UpdatedAt))
		drawStringPc(85, row, style, lp.Fee)
		drawString(w-10, row, style, lp.CommitmentAmount)
		row++

		if row == h-2 {
			if count+1 != len(lps) {
				drawString((w/2)-1, row, whiteStyle, "...")
			}
			break
		}
	}
}

func drawTime() {
	now := time.Now()
	w, _ := ts.Size()
	text := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	drawString(w-8, 0, whiteStyle, text)
}
