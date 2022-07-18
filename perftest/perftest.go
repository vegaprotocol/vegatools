package perftest

import (
	"flag"
	"fmt"
	"time"

	"google.golang.org/grpc"

	datanode "code.vegaprotocol.io/protos/data-node/api/v1"
)

var (
	//	coreService api.CoreServiceClient
	dataNode datanode.TradingDataServiceClient

	// Information about all the users we can use to send orders
	users []UserDetails

	// All the assets knows in the network
	assets map[string]string = map[string]string{}
)

func connectToDataNode(dataNodeAddr string) error {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("Failed to connect to the datanode gRPC port: ", err)
	}

	dataNode = datanode.NewTradingDataServiceClient(connection)

	// load in all the assets
	for len(assets) == 0 {
		getAssets()
		time.Sleep(time.Second * 3)
	}

	return nil
}

func depositTokens(newKeys int, faucetURL string) error {
	if newKeys > 0 {
		for index, user := range users {
			if index > 3 {
				break
			}
			sendVegaTokens(user.pubKey)
			time.Sleep(time.Second * 1)
		}
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

// Run is the main function of `perftest` package
func Run(dataNodeAddr, walletURL, faucetURL string, commandsPerSecond, runtimeSeconds int) error {
	flag.Parse()

	if len(dataNodeAddr) <= 0 {
		return fmt.Errorf("error: missing datanode grpc server address")
	}

	// Connect to data node and check it's working
	err := connectToDataNode(dataNodeAddr)
	if err != nil {
		fmt.Println("Failed to connect to data node: ", err)
	}

	// Create a set of users
	newKeys, err := createOrLoadWallets(walletURL, 10)

	// Send some tokens to any newly created users
	depositTokens(newKeys, faucetURL)

	return nil
}
