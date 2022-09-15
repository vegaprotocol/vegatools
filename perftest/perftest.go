package perftest

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"

	datanode "code.vegaprotocol.io/vega/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/vega/protos/vega"
	commandspb "code.vegaprotocol.io/vega/protos/vega/commands/v1"
)

// Opts hold the command line values
type Opts struct {
	DataNodeAddr      string
	WalletURL         string
	FaucetURL         string
	GanacheURL        string
	CommandsPerSecond int
	RuntimeSeconds    int
	UserCount         int
}

type perfLoadTesting struct {
	// Information about all the users we can use to send orders
	users []UserDetails

	dataNode dnWrapper

	wallet walletWrapper
}

func (p *perfLoadTesting) connectToDataNode(dataNodeAddr string) (map[string]string, error) {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return nil, fmt.Errorf("failed to connect to the datanode gRPC port: %w ", err)
	}

	dn := datanode.NewTradingDataServiceClient(connection)

	p.dataNode = dnWrapper{dataNode: dn,
		wallet: p.wallet}

	// load in all the assets
	for {
		assets, err := p.dataNode.getAssets()
		if err != nil {
			return nil, err
		}
		if len(assets) != 0 {
			return assets, nil
		}
		time.Sleep(time.Second * 1)
	}
}

func (p *perfLoadTesting) CreateUsers(userCount int) error {
	var err error
	p.users, err = p.wallet.CreateOrLoadWallets(userCount)
	return err
}

func (p *perfLoadTesting) depositTokens(assets map[string]string, faucetURL, ganacheURL string) error {
	for index, user := range p.users {
		if index > 3 {
			break
		}
		sendVegaTokens(user.pubKey, ganacheURL)
		time.Sleep(time.Second * 1)
	}

	for _, user := range p.users {
		asset := assets["fUSDC"]
		for amount, _ := p.dataNode.getAssetsPerUser(user.pubKey, asset); amount <= 100000000; {
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Second * 1)
			amount, _ = p.dataNode.getAssetsPerUser(user.pubKey, asset)
		}
	}
	return nil
}

func (p *perfLoadTesting) proposeAndEnactMarket(numberOfMarkets int) ([]string, error) {
	markets := p.dataNode.getMarkets()
	if len(markets) == 0 {
		for i := 0; i < numberOfMarkets; i++ {
			p.wallet.SendNewMarketProposal(i, p.users[0])
			time.Sleep(time.Second * 7)
			propID, err := p.dataNode.getPendingProposalID()
			if err != nil {
				return nil, err
			}
			err = p.dataNode.voteOnProposal(p.users, propID)
			if err != nil {
				return nil, err
			}
			// We have to wait for the market to be enacted
			err = p.dataNode.waitForMarketEnactment(propID, 20)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Second * 6)
		}
	}

	// Move markets out of auction
	markets = p.dataNode.getMarkets()
	if len(markets) >= numberOfMarkets {
		for i := 0; i < len(markets); i++ {
			// Send in a liquidity provision so we can get the market out of auction
			p.wallet.SendLiquidityProvision(p.users[0], markets[i])

			p.wallet.SendOrder(p.users[0], &commandspb.OrderSubmission{MarketId: markets[i],
				Price:       "10010",
				Size:        100,
				Side:        proto.Side_SIDE_SELL,
				Type:        proto.Order_TYPE_LIMIT,
				TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
			p.wallet.SendOrder(p.users[1], &commandspb.OrderSubmission{MarketId: markets[i],
				Price:       "9900",
				Size:        100,
				Side:        proto.Side_SIDE_BUY,
				Type:        proto.Order_TYPE_LIMIT,
				TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
			p.wallet.SendOrder(p.users[0], &commandspb.OrderSubmission{MarketId: markets[i],
				Price:       "10000",
				Size:        5,
				Side:        proto.Side_SIDE_BUY,
				Type:        proto.Order_TYPE_LIMIT,
				TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
			p.wallet.SendOrder(p.users[1], &commandspb.OrderSubmission{MarketId: markets[i],
				Price:       "10000",
				Size:        5,
				Side:        proto.Side_SIDE_SELL,
				Type:        proto.Order_TYPE_LIMIT,
				TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
			time.Sleep(time.Second * 5)
		}
	} else {
		return nil, fmt.Errorf("failed to get open market")
	}

	return markets, nil
}

func (p *perfLoadTesting) sendTradingLoad(marketIDs []string, users, ops, runTimeSeconds int) error {
	// Start load testing by sending off lots of orders at a given rate
	userCount := users - 2
	now := time.Now()
	midPrice := int64(10000)
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
		// Pick a random market to send the trade on
		marketID := marketIDs[rand.Intn(len(marketIDs))]
		userOffset := rand.Intn(userCount) + 2
		user := p.users[userOffset]
		choice := rand.Intn(100)
		if choice < 3 {
			// Perform a cancel all
			err := p.wallet.SendCancelAll(user, marketID)
			if err != nil {
				log.Println("Failed to send cancel all", err)
			}
		} else if choice < 15 {
			// Perform a market order to generate some trades
			if choice%2 == 1 {
				err := p.wallet.SendOrder(user, &commandspb.OrderSubmission{MarketId: marketID,
					Size:        3,
					Side:        proto.Side_SIDE_BUY,
					Type:        proto.Order_TYPE_MARKET,
					TimeInForce: proto.Order_TIME_IN_FORCE_IOC})
				if err != nil {
					log.Println("Failed to send market buy order", err)
				}
			} else {
				err := p.wallet.SendOrder(user, &commandspb.OrderSubmission{MarketId: marketID,
					Size:        3,
					Side:        proto.Side_SIDE_SELL,
					Type:        proto.Order_TYPE_MARKET,
					TimeInForce: proto.Order_TIME_IN_FORCE_IOC})
				if err != nil {
					log.Println("Failed to send market sell order", err)
				}
			}
		} else if choice < 95 {
			// Insert a new order to fill up the book
			priceOffset := rand.Int63n(40) - 20
			if priceOffset > 0 {
				// Send a sell
				err := p.wallet.SendOrder(user, &commandspb.OrderSubmission{MarketId: marketID,
					Price:       fmt.Sprint((midPrice - 1) + priceOffset),
					Size:        1,
					Side:        proto.Side_SIDE_SELL,
					Type:        proto.Order_TYPE_LIMIT,
					TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
				if err != nil {
					log.Println("Failed to send non crossing random limit sell order", err)
				}
			} else {
				// Send a buy
				err := p.wallet.SendOrder(user, &commandspb.OrderSubmission{MarketId: marketID,
					Price:       fmt.Sprint((midPrice + 1) + priceOffset),
					Size:        1,
					Side:        proto.Side_SIDE_BUY,
					Type:        proto.Order_TYPE_LIMIT,
					TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
				if err != nil {
					log.Println("Failed to send non crossing random limit buy order", err)
				}
			}
		} else {
			midPrice = midPrice + (rand.Int63n(20) - 10)
			if midPrice < 9500 {
				midPrice = 9600
			}
			if midPrice > 10500 {
				midPrice = 10400
			}
		}
		count++

		actualDiffSeconds := time.Since(now).Seconds()
		wantedDiffSeconds := float64(count) / opsScale

		// See if we are sending quicker than we should
		if actualDiffSeconds < wantedDiffSeconds {
			delayMillis := (wantedDiffSeconds - actualDiffSeconds) * 1000
			if delayMillis > 10 {
				time.Sleep(time.Millisecond * time.Duration(delayMillis))
				delays++
			}
			actualDiffSeconds = time.Since(now).Seconds()
		}

		if actualDiffSeconds >= 1 {
			fmt.Printf("\rSending load transactions...[%d/%d] %dcps  ", i, numberOfTransactions, count)
			count = 0
			delays = 0
			now = time.Now()
		}
	}
	fmt.Printf("\rSending load transactions...")
	return nil
}

// Run is the main function of `perftest` package
func Run(opts Opts) error {
	flag.Parse()

	plt := perfLoadTesting{wallet: walletWrapper{walletURL: opts.WalletURL}}

	fmt.Print("Connecting to data node...")
	if len(opts.DataNodeAddr) <= 0 {
		fmt.Println("FAILED")
		return fmt.Errorf("error: missing datanode grpc server address")
	}

	// Connect to data node and check it's working
	assets, err := plt.connectToDataNode(opts.DataNodeAddr)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Create a set of users
	fmt.Print("Creating users...")
	err = plt.CreateUsers(opts.UserCount)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send some tokens to any newly created users
	fmt.Print("Depositing tokens and assets...")
	err = plt.depositTokens(assets, opts.FaucetURL, opts.GanacheURL)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send in a proposal to create a new market and vote to get it through
	fmt.Print("Proposing and voting in new market...")
	marketIDs, err := plt.proposeAndEnactMarket(1)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send off a controlled amount of orders and cancels
	fmt.Print("Sending load transactions...")
	err = plt.sendTradingLoad(marketIDs, opts.UserCount, opts.CommandsPerSecond, opts.RuntimeSeconds)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete                      ")

	return nil
}
