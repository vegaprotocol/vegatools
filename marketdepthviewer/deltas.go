package marketdepthviewer

import (
	"context"
	"io"
	"log"
	"time"

	api "code.vegaprotocol.io/vega/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/vega/protos/vega"
)

func getMarketDepthSnapshot(dataclient api.TradingDataServiceClient) error {
	req := &api.MarketDepthRequest{
		MarketId: market.Id,
	}

	resp, err := dataclient.MarketDepth(context.Background(), req)

	if err != nil {
		return err
	}

	// Save the snapshot so we can update it later
	for _, pl := range resp.Buy {
		book.buys[pl.Price] = pl
	}

	for _, pl := range resp.Sell {
		book.sells[pl.Price] = pl
	}
	book.seqNum = resp.SequenceNumber
	return nil
}

func subscribeToMarketDepthUpdates(dataclient api.TradingDataServiceClient) {
	req := &api.MarketDepthUpdatesSubscribeRequest{
		MarketId: market.Id,
	}
	stream, err := dataclient.MarketDepthUpdatesSubscribe(context.Background(), req)
	if err != nil {
		log.Fatalln("Failed to subscribe to trades: ", err)
	}

	// Run in background and process messages
	go processMarketDepthUpdates(stream)

	// Run a background process to make sure we display all updates
	go processBackgroundDisplay()
}

func processMarketDepthUpdates(stream api.TradingDataService_MarketDepthUpdatesSubscribeClient) {
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

		if len(resp.Update.Buy) == 0 && len(resp.Update.Sell) == 0 {
			continue
		}

		if book.seqNum == 0 {
			continue
		}

		if resp.Update.PreviousSequenceNumber != book.seqNum {
			continue
		}
		updateMarketDepthUpdates(resp.Update)
	}
}

func updateMarketDepthUpdates(update *proto.MarketDepthUpdate) {
	for _, pl := range update.Buy {
		if pl.NumberOfOrders == 0 {
			// Remove price level
			delete(book.buys, pl.Price)
		} else {
			// Update price level
			book.buys[pl.Price] = pl
		}
	}
	for _, pl := range update.Sell {
		if pl.NumberOfOrders == 0 {
			// Remove price level
			delete(book.sells, pl.Price)
		} else {
			// Update price level
			book.sells[pl.Price] = pl
		}
	}

	dirty = true
	book.seqNum = update.SequenceNumber

	// If we have already updated in the last second
	if time.Now().After(lastRedraw.Add(time.Second)) {
		drawMarketDepth()
	}
}

func processBackgroundDisplay() {
	for {
		if dirty && time.Now().After(lastRedraw.Add(time.Second)) {
			drawMarketDepth()
		}
		time.Sleep(time.Second)
	}
}
