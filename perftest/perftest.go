package perftest

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc"

	datanode "code.vegaprotocol.io/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/protos/vega"
)

type PerfTestOpts struct {
	DataNodeAddr      string
	WalletURL         string
	FaucetURL         string
	GanacheURL        string
	CommandsPerSecond int
	RuntimeSeconds    int
	UserCount         int
}

var (
	dataNode dnWrapper
	wallet WalletWrapper

	// Information about all the users we can use to send orders
	users []UserDetails
)

func connectToDataNode(dataNodeAddr string) (map[string]string, error) {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return nil, fmt.Errorf("failed to connect to the datanode gRPC port: %w ", err)
	}

	dn := datanode.NewTradingDataServiceClient(connection)

	dataNode = dnWrapper{dataNode: dn,
		wallet: wallet}

	// load in all the assets
	for {
		assets, err := dataNode.getAssets()
		if err != nil {
			return nil, err
		}
		if len(assets) != 0 {
			return assets, nil
		}
		time.Sleep(time.Second * 1)
	}
}

func depositTokens(assets map[string]string, faucetURL, ganacheURL string) error {
	for index, user := range users {
		if index > 3 {
			break
		}
		sendVegaTokens(user.pubKey, ganacheURL)
		time.Sleep(time.Second * 1)
	}

	for _, user := range users {
		asset := assets["fUSDC"]
		for amount, _ := dataNode.getAssetsPerUser(user.pubKey, asset); amount <= 100000000; {
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Second * 1)
			amount, _ = dataNode.getAssetsPerUser(user.pubKey, asset)
		}
	}
	return nil
}

func proposeAndEnactMarket() (string, error) {
	markets := dataNode.getMarkets()
	if len(markets) == 0 {
		wallet.SendNewMarketProposal(0)
		time.Sleep(time.Second * 5)
		propID, err := dataNode.getPendingProposalID()
		if err != nil {
			return "", err
		}
		dataNode.VoteOnProposal(propID)
		time.Sleep(time.Second * 10)
	}

	// Move all markets out of auction
	markets = dataNode.getMarkets()
	if len(markets) > 0 {
		wallet.SendOrder(OrderDetails{markets[0], 0, 10010, -100, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
		wallet.SendOrder(OrderDetails{markets[0], 1, 9900, +100, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
		wallet.SendOrder(OrderDetails{markets[0], 0, 10000, +5, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
		wallet.SendOrder(OrderDetails{markets[0], 1, 10000, -5, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
	} else {
		return "", fmt.Errorf("failed to get open market")
	}
	time.Sleep(time.Second * 5)

	return markets[0], nil
}

func sendTradingLoad(marketID string, users, ops, runTimeSeconds int) error {
	// Start load testing by sending off lots of orders at a given rate
	userCount := users - 2
	now := time.Now()
	count := 0
	delays := 0
	ordersPerSecond := ops
	opsScale := 1.0
	if ordersPerSecond > 1 {
		opsScale = float64(ordersPerSecond - 1)
	}
	// Work out how many orders we need for 10 minute run
	numberOfTransactions := runTimeSeconds * ordersPerSecond
	for i := 0; i < numberOfTransactions; i++ {
		user := rand.Intn(userCount) + 2
		choice := rand.Intn(100)
		if choice < 2 {
			// Perform a cancel all
			wallet.SendCancelAll(user, marketID)
		} else if choice < 7 {
			// Perform a market order to generate some trades
			if choice%2 == 1 {
				wallet.SendOrder(OrderDetails{ marketID, user, 0, 3, proto.Order_TYPE_MARKET, proto.Order_TIME_IN_FORCE_IOC, 0})
			} else {
				wallet.SendOrder(OrderDetails{ marketID, user, 0, -3, proto.Order_TYPE_MARKET, proto.Order_TIME_IN_FORCE_IOC, 0})
			}
		} else {
			// Insert a new order to fill up the book
			priceOffset := rand.Int63n(12) - 6
			if priceOffset > 0 {
				// Send a sell
				wallet.SendOrder(OrderDetails{ marketID, user, int64(10000+user), -1, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
			} else {
				// Send a buy
				wallet.SendOrder(OrderDetails{ marketID, user, int64(9999-user), 1, proto.Order_TYPE_LIMIT, proto.Order_TIME_IN_FORCE_GTC, 0})
			}
		}
		count++

		newNow := time.Now()
		actualDiffSeconds := newNow.Sub(now).Seconds()
		wantedDiffSeconds := float64(count) / opsScale

		// See if we are sending quicker than we should
		if actualDiffSeconds < wantedDiffSeconds {
			delayMillis := (wantedDiffSeconds - actualDiffSeconds) * 1000
			if delayMillis > 10 {
				time.Sleep(time.Millisecond * time.Duration(delayMillis))
				delays++
			}

			newNow = time.Now()
			actualDiffSeconds = newNow.Sub(now).Seconds()
		}

		if actualDiffSeconds >= 1 {
			fmt.Printf("\rSending load transactions...[%d/%d] %dcps  ", i, numberOfTransactions, count)
			count = 0
			delays = 0
			now = newNow
		}
	}
	fmt.Printf("\rSending load transactions...")
	return nil
}

// Run is the main function of `perftest` package
func Run(opts PerfTestOpts) error {
	flag.Parse()

	wallet = WalletWrapper{walletURL: opts.WalletURL}

	fmt.Print("Connecting to data node...")
	if len(opts.DataNodeAddr) <= 0 {
		fmt.Println("FAILED")
		return fmt.Errorf("error: missing datanode grpc server address")
	}

	// Connect to data node and check it's working
	assets, err := connectToDataNode(opts.DataNodeAddr)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Create a set of users
	fmt.Print("Creating users...")
	err = wallet.CreateOrLoadWallets(opts.UserCount)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send some tokens to any newly created users
	fmt.Print("Depositing tokens and assets...")
	err = depositTokens(assets, opts.FaucetURL, opts.GanacheURL)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send in a proposal to create a new market and vote to get it through
	fmt.Print("Proposing and voting in new market...")
	marketID, err := proposeAndEnactMarket()
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send off a controlled amount of orders and cancels
	fmt.Print("Sending load transactions...")
	err = sendTradingLoad(marketID, opts.UserCount, opts.CommandsPerSecond, opts.RuntimeSeconds)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete                      ")

	return nil
}
