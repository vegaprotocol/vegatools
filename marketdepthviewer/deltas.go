package marketdepthviewer

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	api "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	proto "code.vegaprotocol.io/vega/protos/vega"
)

func (m *mdv) getMarketDepthSnapshot(dataclient api.TradingDataServiceClient) error {
	req := &api.GetLatestMarketDepthRequest{
		MarketId: m.market.Id,
	}
	resp, err := dataclient.GetLatestMarketDepth(context.Background(), req)
	if err != nil {
		return err
	}

	// Save the snapshot so we can update it later
	for _, pl := range resp.Buy {
		m.book.buys[pl.Price] = pl
	}

	for _, pl := range resp.Sell {
		m.book.sells[pl.Price] = pl
	}
	m.book.seqNum = resp.SequenceNumber
	return nil
}

func (m *mdv) subscribeToMarketDepthUpdates(dataclient api.TradingDataServiceClient) error {
	req := &api.ObserveMarketsDepthUpdatesRequest{
		MarketIds: []string{m.market.Id},
	}
	stream, err := dataclient.ObserveMarketsDepthUpdates(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to trades: %w", err)
	}

	// Run in background and process messages
	go m.processMarketDepthUpdates(stream)

	// Run a background process to make sure we display all updates
	go m.processBackgroundDisplay()

	return nil
}

func (m *mdv) processMarketDepthUpdates(stream api.TradingDataService_ObserveMarketsDepthUpdatesClient) {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("orders: stream closed by server err: ", err)
			break
		}
		if err != nil {
			log.Println("orders: stream closed err: ", err)
			break
		}

		if len(resp.Update) == 0 {
			continue
		}

		if m.book.seqNum == 0 {
			continue
		}

		for _, md := range resp.Update {
			if md.PreviousSequenceNumber != m.book.seqNum {
				continue
			}
			if len(md.Buy) == 0 && len(md.Sell) == 0 {
				continue
			}
			m.updateMarketDepthUpdates(md)
		}
	}
}

func (m *mdv) updateMarketDepthUpdates(update *proto.MarketDepthUpdate) {
	m.displayMutex.Lock()
	for _, pl := range update.Buy {
		if pl.NumberOfOrders == 0 {
			// Remove price level
			delete(m.book.buys, pl.Price)
		} else {
			// Update price level
			m.book.buys[pl.Price] = pl
		}
	}
	for _, pl := range update.Sell {
		if pl.NumberOfOrders == 0 {
			// Remove price level
			delete(m.book.sells, pl.Price)
		} else {
			// Update price level
			m.book.sells[pl.Price] = pl
		}
	}
	m.dirty = true
	m.book.seqNum = update.SequenceNumber
	m.displayMutex.Unlock()

	// If we have already updated in the last second
	if time.Now().After(m.lastRedraw.Add(time.Second)) {
		m.drawMarketDepth()
	}
}

func (m *mdv) processBackgroundDisplay() {
	for {
		m.displayMutex.Lock()
		shouldRedraw := m.dirty && time.Now().After(m.lastRedraw.Add(time.Second))
		m.displayMutex.Unlock()
		if shouldRedraw {
			m.drawMarketDepth()
		}
		time.Sleep(time.Second)
	}
}
