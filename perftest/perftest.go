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

var (
	// coreService api.CoreServiceClient
	dataNode datanode.TradingDataServiceClient

	// Information about all the users we can use to send orders
	users []UserDetails

	// All the assets knows in the network
	assets map[string]string = map[string]string{}

	// Store the wallet URL for use later
	savedWalletURL string
)

func connectToDataNode(dataNodeAddr string) error {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("failed to connect to the datanode gRPC port: %w ", err)
	}

	dataNode = datanode.NewTradingDataServiceClient(connection)

	// load in all the assets
	for {
		err = getAssets()
		if err != nil {
			return err
		}
		if len(assets) != 0 {
			return nil
		}
		time.Sleep(time.Second * 1)
	}
}

func depositTokens(newKeys int, faucetURL, ganacheURL string) error {
	for index, user := range users {
		if index > 3 {
			break
		}
		sendVegaTokens(user.pubKey, ganacheURL)
		time.Sleep(time.Second * 1)
	}

	for _, user := range users {
		asset := assets["fUSDC"]
		for amount, _ := getAssetsPerUser(user.pubKey, asset); amount <= 100000000; {
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Second * 1)
			amount, _ = getAssetsPerUser(user.pubKey, asset)
		}
	}
	return nil
}

func proposeAndEnactMarket() (string, error) {
	markets := getMarkets()
	if len(markets) == 0 {
		sendNewMarketProposal(0)
		time.Sleep(time.Second * 5)
		propID, err := getPendingProposalID()
		if err != nil {
			return "", err
		}
		voteOnProposal(propID)
		time.Sleep(time.Second * 10)
	}

	// Move all markets out of auction
	markets = getMarkets()
	if len(markets) > 0 {
		sendOrder(markets[0], 0, 10010, -100, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
		sendOrder(markets[0], 1, 9900, +100, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
		sendOrder(markets[0], 0, 10000, +5, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
		sendOrder(markets[0], 1, 10000, -5, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
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
			sendCancelAll(user, marketID)
		} else if choice < 7 {
			// Perform a market order to generate some trades
			if choice%2 == 1 {
				sendOrder(marketID, user, 0, 3, "MARKET", proto.Order_TIME_IN_FORCE_IOC, 0)
			} else {
				sendOrder(marketID, user, 0, -3, "MARKET", proto.Order_TIME_IN_FORCE_IOC, 0)
			}
		} else {
			// Insert a new order to fill up the book
			priceOffset := rand.Int63n(12) - 6
			if priceOffset > 0 {
				// Send a sell
				sendOrder(marketID, user, int64(10000+user), -1, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
			} else {
				// Send a buy
				sendOrder(marketID, user, int64(9999-user), 1, "LIMIT", proto.Order_TIME_IN_FORCE_GTC, 0)
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
func Run(dataNodeAddr, walletURL, faucetURL, ganacheURL string, commandsPerSecond, runtimeSeconds, userCount int) error {
	flag.Parse()

	savedWalletURL = walletURL

	fmt.Print("Connecting to data node...")
	if len(dataNodeAddr) <= 0 {
		fmt.Println("FAILED")
		return fmt.Errorf("error: missing datanode grpc server address")
	}

	// Connect to data node and check it's working
	err := connectToDataNode(dataNodeAddr)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Create a set of users
	fmt.Print("Creating users...")
	newKeys, err := createOrLoadWallets(walletURL, userCount)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send some tokens to any newly created users
	fmt.Print("Depositing tokens and assets...")
	err = depositTokens(newKeys, faucetURL, ganacheURL)
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
	err = sendTradingLoad(marketID, userCount, commandsPerSecond, runtimeSeconds)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete                      ")

	return nil
}
