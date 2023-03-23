package grpcapibenchmark

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"code.vegaprotocol.io/vega/datanode/ratelimit"
	v2 "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/spf13/cobra"
)

var (
	// ListOrdersCmd is the command that will be run when running a ListOrders benchmark
	ListOrdersCmd = &cobra.Command{
		Use:   "ListOrders",
		Short: "Test orders API",
		Long:  "Run tests against the orders APIs",
		Run:   listOrders,
	}

	urls       []string
	marketID   string
	partyID    string
	reference  string
	timeout    time.Duration
	iterations int
	workers    int
	requests   int
	startDate  string
	endDate    string
)

func init() {
	ListOrdersCmd.Flags().StringVarP(&marketID, "market", "m", "", "UUID of market to query orders for")
	ListOrdersCmd.Flags().StringVarP(&partyID, "party", "p", "", "UUID of Party to query orders for")
	ListOrdersCmd.Flags().StringVarP(&reference, "reference", "r", "", "Status to query orders for")
	ListOrdersCmd.Flags().StringVarP(&startDate, "start-date", "s", "", "Start date for the date range to use in the query")
	ListOrdersCmd.Flags().StringVarP(&endDate, "end-date", "e", "", "Start date for the date range to use in the query")

	if startDate != "" {
		_, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			log.Fatalf("could not parse start date, please use RFC3339 format, %v", err)
		}
	}

	if endDate != "" {
		_, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			log.Fatalf("could not parse end date, please use RFC3339 format, %v", err)
		}
	}
}

func listOrders(command *cobra.Command, args []string) {
	parsePersistentFlags(command)

	log.Printf("Benchmarking ListOrders for: %v", urls)
	log.Print(", Iteration, Host, mean, median, latency, latency std dev")

	for _, url := range urls {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()), ratelimit.WithSecret())
		if err != nil {
			log.Printf("could not connect to %s, %v", url, err)
			continue
		}
		client := v2.NewTradingDataServiceClient(conn)

		pinger, err := probing.NewPinger(getHost(url))
		if err != nil {
			log.Printf("could not create pinger, %v", err)
			continue
		}

		pinger.Count = 5
		if err = pinger.Run(); err != nil {
			log.Printf("could not ping %s, %v", url, err)
			continue
		}

		rtts := pinger.Statistics().AvgRtt
		stdDev := pinger.Statistics().StdDevRtt

		for i := 0; i < iterations; i++ {
			reqCh := make(chan struct{})
			resultCh := make(chan time.Duration)
			doneCh := make(chan struct{})

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			for i := 0; i < workers; i++ {
				go worker(ctx, client, listOrdersRequest, reqCh, resultCh, doneCh)
			}

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				for i := 0; i < requests; i++ {
					reqCh <- struct{}{}
					// slight pause to avoid hitting the rate limiter
					time.Sleep(time.Millisecond * 5)
				}
				close(reqCh)
			}()

			metrics := make([]time.Duration, 0)

			go func() {
				count := 0
				for {
					select {
					case m := <-resultCh:
						count++
						if m > 0 {
							metrics = append(metrics, m)
						}
					case <-doneCh:
						close(resultCh)
						wg.Done()
						return
					}
				}
			}()
			wg.Wait()
			mean := mean(metrics)
			median := median(metrics)

			log.Printf(", %d, %s, %v, %v, %v, %v", i+1, url, mean.Milliseconds(), median.Milliseconds(), rtts, stdDev)

			cancel()
		}
	}
}

func parsePersistentFlags(command *cobra.Command) {
	var err error
	urls, err = command.Flags().GetStringSlice("urls")
	if err != nil {
		panic(err)
	}

	timeout, err = command.Flags().GetDuration("timeout")
	if err != nil {
		panic(err)
	}

	iterations, err = command.Flags().GetInt("iterations")
	if err != nil {
		panic(err)
	}

	workers, err = command.Flags().GetInt("workers")
	if err != nil {
		panic(err)
	}

	requests, err = command.Flags().GetInt("requests")
	if err != nil {
		panic(err)
	}
}

func listOrdersRequest(ctx context.Context, client v2.TradingDataServiceClient) time.Duration {
	start := time.Now()
	var market *string
	var party *string
	var ref *string

	if marketID != "" {
		market = &marketID
	}

	if partyID != "" {
		party = &partyID
	}

	if reference != "" {
		ref = &reference
	}

	var dateRangeStart, dateRangeEnd *int64

	if startDate != "" {
		s, _ := time.Parse(time.RFC3339, startDate)
		startNanos := s.UnixNano()
		dateRangeStart = &startNanos
	}

	if endDate != "" {
		e, _ := time.Parse(time.RFC3339, endDate)
		endNanos := e.UnixNano()
		dateRangeEnd = &endNanos
	}

	var dateRange *v2.DateRange
	if dateRangeStart != nil || dateRangeEnd != nil {
		dateRange = &v2.DateRange{
			StartTimestamp: dateRangeStart,
			EndTimestamp:   dateRangeEnd,
		}
	}

	if _, err := client.ListOrders(ctx, &v2.ListOrdersRequest{
		PartyId:   party,
		MarketId:  market,
		Reference: ref,
		DateRange: dateRange,
	}); err != nil {
		return 0
	}

	elapsed := time.Since(start)
	return elapsed
}

func getHost(uri string) string {
	host, _, _ := net.SplitHostPort(uri)
	return host
}
