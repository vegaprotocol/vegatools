package perftest

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	datanode "code.vegaprotocol.io/vega/protos/data-node/api/v2"
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
	MarketCount       int
	Voters            int
	MoveMid           bool
	LPOrdersPerSide   int
	BatchSize         int
}

type perfLoadTesting struct {
	// Information about all the users we can use to send orders
	users []UserDetails

	dataNode dnWrapper

	wallet walletWrapper
}

func (p *perfLoadTesting) connectToDataNode(dataNodeAddr string) (map[string]string, error) {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (p *perfLoadTesting) depositTokens(assets map[string]string, faucetURL, ganacheURL string, voters int) error {
	for index, user := range p.users {
		if index >= voters {
			break
		}
		sendVegaTokens(user.pubKey, ganacheURL)
		time.Sleep(time.Second * 1)
	}

	// If the first user has not tokens, top everyone up
	// quickly without checking if they need it
	asset := assets["fUSDC"]
	amount, _ := p.dataNode.getAssetsPerUser(p.users[0].pubKey, asset)
	if amount == 0 {
		for _, user := range p.users {
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Millisecond * 50)
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Millisecond * 50)
		}
	}

	for _, user := range p.users {
		for amount, _ := p.dataNode.getAssetsPerUser(user.pubKey, asset); amount <= 100000000; {
			topUpAsset(faucetURL, user.pubKey, asset, 100000000)
			time.Sleep(time.Second * 1)
			amount, _ = p.dataNode.getAssetsPerUser(user.pubKey, asset)
		}
	}

	time.Sleep(time.Second * 5)
	return nil
}

func (p *perfLoadTesting) checkNetworkLimits(opts Opts) error {
	// Check the limit of the number of orders per side in the LP shape
	networkParam, err := p.dataNode.getNetworkParam("market.liquidityProvision.shapes.maxSize")
	if err != nil {
		fmt.Println("Failed to get LP maximum shape size")
		return err
	}
	maxLPShape, _ := strconv.ParseInt(networkParam, 0, 32)

	if opts.LPOrdersPerSide > int(maxLPShape) {
		return fmt.Errorf("supplied lp size greater than network param (%d>%d)", opts.LPOrdersPerSide, maxLPShape)
	}

	// Check the maximum number of orders in a batch
	networkParam, err = p.dataNode.getNetworkParam("spam.protection.max.batchSize")
	if err != nil {
		fmt.Println("Failed to get maximum order batch size")
		return err
	}
	maxBatchSize, _ := strconv.ParseInt(networkParam, 0, 32)

	if opts.BatchSize > int(maxBatchSize) {
		return fmt.Errorf("supplied order batch size is greater than network param (%d>%d)", opts.BatchSize, maxBatchSize)
	}
	return nil
}

func (p *perfLoadTesting) proposeAndEnactMarket(numberOfMarkets, voters, maxLPShape int) ([]string, error) {
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
			for j := 0; j < voters; j++ {
				p.wallet.SendLiquidityProvision(p.users[j], markets[i], maxLPShape)
			}

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

func (p *perfLoadTesting) sendTradingLoad(marketIDs []string, users, ops, runTimeSeconds int, moveMid bool) error {
	// Start load testing by sending off lots of orders at a given rate
	userCount := users - 2
	now := time.Now()
	midPrice := int64(10000)
	transactionCount := 0
	delays := 0
	transactionsPerSecond := ops
	opsScale := 1.0
	if transactionsPerSecond > 1 {
		opsScale = float64(transactionsPerSecond - 1)
	}
	// Work out how many transactions we need for the length of the run
	numberOfTransactions := runTimeSeconds * transactionsPerSecond
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

			if moveMid {
				// Move the midprice around as well
				midPrice = midPrice + (rand.Int63n(3) - 1)
				if midPrice < 9500 {
					midPrice = 9505
				}
				if midPrice > 10500 {
					midPrice = 10495
				}
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
		} else {
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
		}
		transactionCount++

		actualDiffSeconds := time.Since(now).Seconds()
		wantedDiffSeconds := float64(transactionCount) / opsScale

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
			fmt.Printf("\rSending load transactions...[%d/%d] %dcps  ", i, numberOfTransactions, transactionCount)
			transactionCount = 0
			delays = 0
			now = time.Now()
		}
	}
	fmt.Printf("\rSending load transactions...")
	return nil
}

func (p *perfLoadTesting) sendBatchTradingLoad(marketIDs []string, users, ops, runTimeSeconds, batchSize int, moveMid bool) error {
	userCount := users - 2
	now := time.Now()
	midPrice := int64(10000)
	transactionCount := 0
	totalTransactionCount := 0

	for i := 0; i < runTimeSeconds; i++ {
		// Pick a random market to send the trade on
		marketID := marketIDs[rand.Intn(len(marketIDs))]
		userOffset := rand.Intn(userCount) + 2
		user := p.users[userOffset]

		cancels := []*commandspb.OrderCancellation{}
		amends := []*commandspb.OrderAmendment{}
		orders := []*commandspb.OrderSubmission{}

		batchCount := 0
		// Now process transactions for this user and market in a batch
		for j := 0; j < ops; j++ {
			choice := rand.Intn(100)
			if choice < 3 {
				cancels = append(cancels, &commandspb.OrderCancellation{
					MarketId: marketID,
				})
				if moveMid {
					// Move the midprice around as well
					midPrice = midPrice + (rand.Int63n(3) - 1)
					if midPrice < 9500 {
						midPrice = 9505
					}
					if midPrice > 10500 {
						midPrice = 10495
					}
				}
			} else if choice < 15 {
				// Perform a market order to generate some trades
				if choice%2 == 1 {
					orders = append(orders, &commandspb.OrderSubmission{MarketId: marketID,
						Size:        3,
						Side:        proto.Side_SIDE_BUY,
						Type:        proto.Order_TYPE_MARKET,
						TimeInForce: proto.Order_TIME_IN_FORCE_IOC})
				} else {
					orders = append(orders, &commandspb.OrderSubmission{MarketId: marketID,
						Size:        3,
						Side:        proto.Side_SIDE_SELL,
						Type:        proto.Order_TYPE_MARKET,
						TimeInForce: proto.Order_TIME_IN_FORCE_IOC})
				}
			} else {
				// Insert a new order to fill up the book
				priceOffset := rand.Int63n(40) - 20
				if priceOffset > 0 {
					// Send a sell
					orders = append(orders, &commandspb.OrderSubmission{MarketId: marketID,
						Price:       fmt.Sprint((midPrice - 1) + priceOffset),
						Size:        1,
						Side:        proto.Side_SIDE_SELL,
						Type:        proto.Order_TYPE_LIMIT,
						TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
				} else {
					// Send a buy
					orders = append(orders, &commandspb.OrderSubmission{MarketId: marketID,
						Price:       fmt.Sprint((midPrice + 1) + priceOffset),
						Size:        1,
						Side:        proto.Side_SIDE_BUY,
						Type:        proto.Order_TYPE_LIMIT,
						TimeInForce: proto.Order_TIME_IN_FORCE_GTC})
				}
			}
			transactionCount++
			batchCount++
			if batchCount == batchSize {
				// Send off the batch and reset everything
				err := p.wallet.SendBatchOrders(user, cancels, amends, orders)
				if err != nil {
					return err
				}
				batchCount = 0
				marketID = marketIDs[rand.Intn(len(marketIDs))]
				userOffset = rand.Intn(userCount) + 2
				user = p.users[userOffset]

				cancels = cancels[:0]
				amends = amends[:0]
				orders = orders[:0]
			}
		}

		// Now send off all the commands in a batch
		err := p.wallet.SendBatchOrders(user, cancels, amends, orders)
		if err != nil {
			return err
		}

		timeUsed := time.Since(now).Seconds()

		// Add in a delay to keep us processing at a per second rate
		if timeUsed < 1.0 {
			milliSecondsLeft := int((1.0 - timeUsed) * 1000.0)
			time.Sleep(time.Millisecond * time.Duration(milliSecondsLeft))
		}
		totalTransactionCount += transactionCount
		fmt.Printf("\rSending load transactions...[%d/%d] %dcps  ", totalTransactionCount, ops*runTimeSeconds, transactionCount)
		transactionCount = 0
		now = time.Now()
	}
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
	err = plt.depositTokens(assets, opts.FaucetURL, opts.GanacheURL, opts.Voters)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	err = plt.checkNetworkLimits(opts)
	if err != nil {
		return err
	}

	// Send in a proposal to create a new market and vote to get it through
	fmt.Print("Proposing and voting in new market...")
	marketIDs, err := plt.proposeAndEnactMarket(opts.MarketCount, opts.Voters, opts.LPOrdersPerSide)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("Complete")

	// Send off a controlled amount of orders and cancels
	if opts.BatchSize > 0 {
		fmt.Print("Sending load transactions...")
		err = plt.sendBatchTradingLoad(marketIDs, opts.UserCount, opts.CommandsPerSecond, opts.RuntimeSeconds, opts.BatchSize, opts.MoveMid)
		if err != nil {
			fmt.Println("FAILED")
			return err
		}
	} else {
		fmt.Print("Sending load transactions...")
		err = plt.sendTradingLoad(marketIDs, opts.UserCount, opts.CommandsPerSecond, opts.RuntimeSeconds, opts.MoveMid)
		if err != nil {
			fmt.Println("FAILED")
			return err
		}
	}
	fmt.Println("Complete                      ")

	return nil
}
