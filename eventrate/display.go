package eventrate

import (
	"fmt"
	"log"
	"sort"

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
	greyStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorGrey)

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
	drawEvents()
	ts.Show()
}

func drawHeaders() {
	w, h := ts.Size()
	drawString(0, 0, whiteStyle, server)
	drawString(w/2, 0, whiteStyle, "EVENT RATE HEADER")
	drawString(w/2, h-1, whiteStyle, "EVENT RATE FOOTER")
}

func drawEvents() {
	w, _ := ts.Size()

	// Turn the historic map into an array we can sort
	events := make([]*counter, 0, len(historicEventCounts))
	for _, counter := range historicEventCounts {
		events = append(events, counter)
	}

	// Sort them using the average rate value
	sort.Slice(events, func(i, j int) bool {
		return events[i].avgEventsPerSecond > events[j].avgEventsPerSecond
	})

	for y := 0; y < len(events); y++ {
		c := events[y]
		numStr := fmt.Sprintf("%d %d %d", c.minEventsPerSecond, c.avgEventsPerSecond, c.maxEventsPerSecond)
		drawString(0, y+2, whiteStyle, events[y].eventType.String())
		drawString(w/2, y+2, greyStyle, numStr)
	}
}
