package perftest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	datanode "code.vegaprotocol.io/vega/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/vega/protos/vega"
	v1 "code.vegaprotocol.io/vega/protos/vega/commands/v1"
)

type dnWrapper struct {
	dataNode datanode.TradingDataServiceClient
	wallet   walletWrapper
}

func (d *dnWrapper) getAssets() (map[string]string, error) {
	request := &datanode.AssetsRequest{}

	response, err := d.dataNode.Assets(context.Background(), request)
	if err != nil {
		return nil, err
	}
	assets := map[string]string{}
	for _, asset := range response.Assets {
		assets[asset.Details.Symbol] = asset.Id
	}
	return assets, nil
}

func (d *dnWrapper) getAssetsPerUser(pubKey, asset string) (int64, error) {
	request := &datanode.PartyAccountsRequest{
		PartyId: pubKey,
		Asset:   asset,
	}

	response, err := d.dataNode.PartyAccounts(context.Background(), request)
	if err != nil {
		return 0, err
	}
	for _, account := range response.Accounts {
		if account.Type == proto.AccountType_ACCOUNT_TYPE_GENERAL {
			return strconv.ParseInt(account.Balance, 10, 64)
		}
	}
	return 0, nil
}

func (d *dnWrapper) getMarkets() []string {
	marketsReq := &datanode.MarketsRequest{}

	response, err := d.dataNode.Markets(context.Background(), marketsReq)
	if err != nil {
		log.Println(err)
		return nil
	}
	marketIDs := []string{}
	for _, market := range response.Markets {
		if market.State != proto.Market_STATE_REJECTED {
			marketIDs = append(marketIDs, market.Id)
		}
	}
	return marketIDs
}

func (d *dnWrapper) getPendingProposalID() (string, error) {
	request := &datanode.GetProposalsRequest{}

	response, err := d.dataNode.GetProposals(context.Background(), request)
	if err != nil {
		log.Println(err)
	}

	for _, proposal := range response.GetData() {
		if proposal.Proposal.State == proto.Proposal_STATE_OPEN {
			return proposal.Proposal.Id, nil
		}
	}
	return "", fmt.Errorf("no pending proposals found")
}

func (d *dnWrapper) waitForMarketEnactment(marketID string, maxWaitSeconds int) error {
	request := &datanode.GetProposalsRequest{}

	for i := 0; i < maxWaitSeconds; i++ {
		response, err := d.dataNode.GetProposals(context.Background(), request)
		if err != nil {
			return err
		}

		for _, proposal := range response.GetData() {
			if proposal.Proposal.State == proto.Proposal_STATE_ENACTED {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Timed out waiting for market to be enacted")
}

func (d *dnWrapper) voteOnProposal(users []UserDetails, propID string) error {
	for i := 0; i < 3; i++ {
		err := d.wallet.SendVote(users[i], &v1.VoteSubmission{ProposalId: propID, Value: proto.Vote_VALUE_YES})
		if err != nil {
			return err
		}
	}
	return nil
}
