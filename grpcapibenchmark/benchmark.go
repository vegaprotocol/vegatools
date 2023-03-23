package grpcapibenchmark

import (
	"context"
	"sort"
	"time"

	v2 "code.vegaprotocol.io/vega/protos/data-node/api/v2"
)

type apiTest func(ctx context.Context, client v2.TradingDataServiceClient) time.Duration

func worker(ctx context.Context, client v2.TradingDataServiceClient, apiTest apiTest, reqCh <-chan struct{}, resultCh chan<- time.Duration, doneCh chan<- struct{}) {
	defer func() {
		doneCh <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-reqCh:
			if !ok {
				return
			}
			elapsed := apiTest(ctx, client)
			resultCh <- elapsed
		}
	}
}

type averagable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | time.Duration
}

func mean[T averagable](values []T) T {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return values[0]
	}

	var total T
	for _, d := range values {
		total += d
	}
	return total / T(len(values))
}

func median[T averagable](values []T) T {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return values[0]
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	middle := len(values) / 2
	if len(values)%2 == 0 {
		return (values[middle-1] + values[middle]) / 2
	}
	return values[middle]
}
