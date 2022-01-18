package withdraw

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "code.vegaprotocol.io/protos/data-node/api/v1"
	"code.vegaprotocol.io/protos/vega"

	"google.golang.org/grpc"
)

// PullWithdraw ...
func PullWithdraw(nodeAddress, outfile string) error {
	conn, err := grpc.Dial(nodeAddress, grpc.WithInsecure())
	if err != nil {
		// Something went wrong
		return fmt.Errorf("failed to connect to the vega gRPC port: %s", err)
	}
	defer conn.Close()
	clt := api.NewTradingDataServiceClient(conn)

	parties, err := getParties(clt)
	if err != nil {
		return err
	}

	withdrawals, err := getWithdrawals(clt, parties)
	if err != nil {
		return err
	}

	out, err := getBundles(clt, withdrawals)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(out)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", string(buf))

	return nil
}

type withdrawalBundle struct {
	AssetSource   string
	Amount        string
	Expiry        int64
	Nonce         string
	Signatures    string
	TargetAddress string
}

type withdrawalAndBundlePair struct {
	Withdrawal *vega.Withdrawal
	Bundle     *withdrawalBundle
}

func getBundles(clt api.TradingDataServiceClient, withdrawals map[string][]*vega.Withdrawal) (map[string][]withdrawalAndBundlePair, error) {
	out := map[string][]withdrawalAndBundlePair{}

	var err error
	for k, v := range withdrawals {
		if len(v) > 0 {
			out[k], err = getPartyWithdrawalAndBundlePairs(clt, v)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func getPartyWithdrawalAndBundlePairs(clt api.TradingDataServiceClient, withdrawals []*vega.Withdrawal) ([]withdrawalAndBundlePair, error) {
	out := []withdrawalAndBundlePair{}

	for _, v := range withdrawals {
		bundle, err := getBundle(clt, v)
		if err != nil {
			continue // possible there's no bundle if withdrawal was invalid
		}

		out = append(out, withdrawalAndBundlePair{
			Withdrawal: v,
			Bundle:     bundle,
		})
	}
	return out, nil
}

func getBundle(clt api.TradingDataServiceClient, w *vega.Withdrawal) (*withdrawalBundle, error) {
	ctx, cfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfunc()
	resp, err := clt.ERC20WithdrawalApproval(ctx, &api.ERC20WithdrawalApprovalRequest{
		WithdrawalId: w.Id,
	})
	if err != nil {
		return nil, err
	}

	return &withdrawalBundle{
		AssetSource:   resp.AssetSource,
		Amount:        resp.Amount,
		Expiry:        resp.Expiry,
		Nonce:         resp.Nonce,
		Signatures:    resp.Signatures,
		TargetAddress: w.GetExt().GetErc20().GetReceiverAddress(),
	}, nil
}

func getWithdrawals(clt api.TradingDataServiceClient, parties []*vega.Party) (map[string][]*vega.Withdrawal, error) {
	withdrawals := map[string][]*vega.Withdrawal{}

	var err error
	for _, v := range parties {
		withdrawals[v.Id], err = getPartyWithdrawals(clt, v.Id)
		if err != nil {
			return nil, err
		}
	}

	return withdrawals, nil
}

func getPartyWithdrawals(clt api.TradingDataServiceClient, party string) ([]*vega.Withdrawal, error) {
	ctx, cfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfunc()
	withdrawals, err := clt.Withdrawals(ctx, &api.WithdrawalsRequest{PartyId: party})
	if err != nil {
		return nil, err
	}

	return withdrawals.Withdrawals, nil
}

func getParties(clt api.TradingDataServiceClient) ([]*vega.Party, error) {
	ctx, cfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfunc()
	parties, err := clt.Parties(ctx, &api.PartiesRequest{})
	if err != nil {
		return nil, err
	}
	return parties.Parties, nil
}
