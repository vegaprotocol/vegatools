package perftest

import (
	"context"
	"log"
	"strconv"

	datanode "code.vegaprotocol.io/protos/data-node/api/v1"
)

func getAssets() {
	request := &datanode.AssetsRequest{}

	response, err := dataNode.Assets(context.Background(), request)
	if err != nil {
		log.Println(err)
		return
	}
	for _, asset := range response.Assets {
		assets[asset.Details.Symbol] = asset.Id
	}
}

func getAssetsPerUser(pubKey, asset string) (int64, error) {
	request := &datanode.PartyAccountsRequest{
		PartyId: pubKey,
		Asset:   asset,
	}

	response, err := dataNode.PartyAccounts(context.Background(), request)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	for _, account := range response.Accounts {
		log.Println(account)
		return strconv.ParseInt(account.Balance, 10, 64)
	}
	return 0, nil
}
