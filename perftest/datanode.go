package perftest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	datanode "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	proto "code.vegaprotocol.io/vega/protos/vega"
)

type dnWrapper struct {
	dataNode datanode.TradingDataServiceClient
	wallet   walletWrapper
}

func (d *dnWrapper) getNetworkParam(param string) (string, error) {
	request := &datanode.GetNetworkParameterRequest{
		Key: param,
	}

	response, err := d.dataNode.GetNetworkParameter(context.Background(), request)
	if err != nil {
		return "", err
	}

	return response.NetworkParameter.Value, nil
}

func (d *dnWrapper) getStake(partyID string) (int64, error) {
	request := &datanode.GetStakeRequest{
		PartyId: partyID,
	}

	response, err := d.dataNode.GetStake(context.Background(), request)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(response.CurrentStakeAvailable, 10, 64)
}

func (d *dnWrapper) getAssets() (map[string]string, error) {
	request := &datanode.ListAssetsRequest{}

	response, err := d.dataNode.ListAssets(context.Background(), request)
	if err != nil {
		return nil, err
	}
	assets := map[string]string{}
	for _, asset := range response.GetAssets().GetEdges() {
		assets[asset.Node.Details.Symbol] = asset.Node.Id
	}
	return assets, nil
}

func (d *dnWrapper) getAssetsPerUser(pubKey, asset string) (int64, error) {
	request := &datanode.ListAccountsRequest{
		Filter: &datanode.AccountFilter{
			AssetId:  asset,
			PartyIds: []string{pubKey},
			AccountTypes: []proto.AccountType{
				proto.AccountType_ACCOUNT_TYPE_GENERAL,
			},
		},
	}

	response, err := d.dataNode.ListAccounts(context.Background(), request)
	if err != nil {
		return 0, err
	}
	for _, account := range response.Accounts.Edges {
		return strconv.ParseInt(account.Node.Balance, 10, 64)
	}
	return 0, nil
}

func (d *dnWrapper) getMarkets() []*proto.Market {
	marketsReq := &datanode.ListMarketsRequest{}

	response, err := d.dataNode.ListMarkets(context.Background(), marketsReq)
	if err != nil {
		log.Println(err)
		return nil
	}
	marketIDs := []*proto.Market{}
	for _, market := range response.Markets.Edges {
		if market.Node.State != proto.Market_STATE_REJECTED {
			marketIDs = append(marketIDs, market.Node)
		}
	}
	return marketIDs
}

func (d *dnWrapper) getPendingProposalID() (string, error) {
	request := &datanode.ListGovernanceDataRequest{}

	response, err := d.dataNode.ListGovernanceData(context.Background(), request)
	if err != nil {
		log.Println(err)
	}

	for _, proposal := range response.Connection.Edges {
		if proposal.Node.Proposal.State == proto.Proposal_STATE_OPEN {
			return proposal.Node.Proposal.Id, nil
		}
	}
	return "", fmt.Errorf("no pending proposals found")
}

func (d *dnWrapper) waitForMarketEnactment(marketID string, maxWaitSeconds int) error {
	request := &datanode.ListGovernanceDataRequest{
		ProposalReference: &marketID,
	}

	for i := 0; i < maxWaitSeconds; i++ {
		response, err := d.dataNode.ListGovernanceData(context.Background(), request)
		if err != nil {
			return err
		}

		for _, proposal := range response.Connection.Edges {
			if proposal.Node.Proposal.State == proto.Proposal_STATE_ENACTED {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timed out waiting for market to be enacted")
}

func (d *dnWrapper) voteOnProposal(users []UserDetails, propID string, voters int) error {
	for i := 0; i < voters; i++ {
		err := d.wallet.SendVote(users[i], propID)
		if err != nil {
			return err
		}
	}
	return nil
}
