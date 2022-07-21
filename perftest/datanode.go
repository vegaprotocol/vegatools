package perftest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	datanode "code.vegaprotocol.io/protos/data-node/api/v1"
	proto "code.vegaprotocol.io/protos/vega"
)

func getAssets() error {
	request := &datanode.AssetsRequest{}

	response, err := dataNode.Assets(context.Background(), request)
	if err != nil {
		return err
	}
	for _, asset := range response.Assets {
		assets[asset.Details.Symbol] = asset.Id
	}
	return nil
}

func getAssetsPerUser(pubKey, asset string) (int64, error) {
	request := &datanode.PartyAccountsRequest{
		PartyId: pubKey,
		Asset:   asset,
	}

	response, err := dataNode.PartyAccounts(context.Background(), request)
	if err != nil {
		return 0, err
	}
	for _, account := range response.Accounts {
		return strconv.ParseInt(account.Balance, 10, 64)
	}
	return 0, nil
}

func getMarkets() []string {
	marketsReq := &datanode.MarketsRequest{}

	response, err := dataNode.Markets(context.Background(), marketsReq)
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

func getPendingProposalID() (string, error) {
	request := &datanode.GetProposalsRequest{}

	response, err := dataNode.GetProposals(context.Background(), request)
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

func voteOnProposal(propID string) error {
	for i := 0; i < 3; i++ {
		err := sendVote(i, propID, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendVote(user int, proposalID string, vote bool) error {
	cmd := `{ "voteSubmission": {
              "proposal_id": "$PROPOSAL_ID",
              "value": "$VOTE"
            },
            "pubKey": "$PUBKEY",
            "propagate" : true
          }`

	cmd = strings.Replace(cmd, "$PROPOSAL_ID", proposalID, 1)
	cmd = strings.Replace(cmd, "$PUBKEY", users[user].pubKey, 1)
	if vote {
		cmd = strings.Replace(cmd, "$VOTE", "VALUE_YES", 1)
	} else {
		cmd = strings.Replace(cmd, "$VOTE", "VALUE_NO", 1)
	}

	err := signSubmitTx(user, cmd)
	if err != nil {
		return err
	}
	return nil
}
