package diff

import (
	"context"
	"sync"

	"code.vegaprotocol.io/vega/libs/crypto"
	dn "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"code.vegaprotocol.io/vega/protos/vega"
	v1 "code.vegaprotocol.io/vega/protos/vega/events/v1"
	"google.golang.org/grpc"
)

type dataNodeClient struct {
	datanode dn.TradingDataServiceClient
}

func newDataNodeClient(dataNodeAddr string) *dataNodeClient {
	connection, err := grpc.Dial(dataNodeAddr, grpc.WithInsecure())
	if err != nil {
		return nil
	}

	return &dataNodeClient{
		datanode: dn.NewTradingDataServiceClient(connection),
	}
}

func (dnc *dataNodeClient) Collect() (*Result, error) {
	res := &Result{}
	var wg sync.WaitGroup
	wg.Add(17)

	errors := make(chan error)

	go func() {
		defer wg.Done()
		var err error
		res.Accounts, err = dnc.listAccounts()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Orders, err = dnc.listOrders()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Markets, err = dnc.listMarkets()
		if err != nil {
			errors <- err
		} else {
			for _, m := range res.Markets {
				var lps []*vega.LiquidityProvision
				lps, err = dnc.listLiquidityProvisions(m.Id)
				if err != nil {
					errors <- err
				}
				res.Lps = append(res.Lps, lps...)
			}
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Parties, err = dnc.listParties()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Limits, err = dnc.getNetworkLimits()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Assets, err = dnc.listAssets()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.VegaTime, err = dnc.getVegaTime()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Delegations, err = dnc.listDelegations()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Epoch, err = dnc.getEpoch()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Nodes, err = dnc.listNodes()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.NetParams, err = dnc.listNetworkParameters()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Proposals, err = dnc.listGovernanceData()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Deposits, err = dnc.listDeposits()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Withdrawals, err = dnc.listWithdrawals()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Positions, err = dnc.listPositions()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Transfers, err = dnc.listTransfers()
		if err != nil {
			errors <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		res.Stake, err = dnc.getStake()
		if err != nil {
			errors <- err
		}
	}()

	var resErr error
	go func() {
		for r := range errors {
			resErr = r
		}
	}()
	wg.Wait()

	return res, resErr
}

func (dnc *dataNodeClient) listAccounts() ([]*dn.AccountBalance, error) {
	accResp, err := dnc.datanode.ListAccounts(context.Background(), &dn.ListAccountsRequest{})
	if err != nil {
		return nil, err
	}
	accounts := make([]*dn.AccountBalance, 0, len(accResp.Accounts.Edges))
	for _, ae := range accResp.Accounts.Edges {
		accounts = append(accounts, ae.Node)
	}
	return accounts, nil
}

func (dnc *dataNodeClient) listOrders() ([]*vega.Order, error) {
	liveOnly := true
	orderResp, err := dnc.datanode.ListOrders(context.Background(), &dn.ListOrdersRequest{
		Filter: &dn.OrderFilter{
			LiveOnly: &liveOnly},
	})
	if err != nil {
		return nil, err
	}
	orders := make([]*vega.Order, 0, len(orderResp.Orders.Edges))
	for _, oe := range orderResp.Orders.Edges {
		if oe.Node.Status != vega.Order_STATUS_PARKED {
			orders = append(orders, oe.Node)
		}
	}
	return orders, nil
}

func (dnc *dataNodeClient) listMarkets() ([]*vega.Market, error) {
	marketResp, err := dnc.datanode.ListMarkets(context.Background(), &dn.ListMarketsRequest{})
	if err != nil {
		return nil, err
	}
	markets := make([]*vega.Market, 0, len(marketResp.Markets.Edges))
	for _, me := range marketResp.Markets.Edges {
		markets = append(markets, me.Node)
	}
	return markets, nil
}

func (dnc *dataNodeClient) listParties() ([]*vega.Party, error) {
	partiesResp, err := dnc.datanode.ListParties(context.Background(), &dn.ListPartiesRequest{})
	if err != nil {
		return nil, err
	}
	parties := make([]*vega.Party, 0, len(partiesResp.Parties.Edges))
	for _, pe := range partiesResp.Parties.Edges {
		parties = append(parties, pe.Node)
	}
	return parties, nil
}

func (dnc *dataNodeClient) getNetworkLimits() (*vega.NetworkLimits, error) {
	limitsResp, err := dnc.datanode.GetNetworkLimits(context.Background(), &dn.GetNetworkLimitsRequest{})
	if err != nil {
		return nil, err
	}
	return limitsResp.Limits, nil
}

func (dnc *dataNodeClient) listAssets() ([]*vega.Asset, error) {
	assetResp, err := dnc.datanode.ListAssets(context.Background(), &dn.ListAssetsRequest{})
	if err != nil {
		return nil, err
	}
	assets := make([]*vega.Asset, 0, len(assetResp.Assets.Edges))
	for _, a := range assetResp.Assets.Edges {
		if a.Node.Status != vega.Asset_STATUS_REJECTED {
			assets = append(assets, a.Node)
		}
	}
	return assets, nil
}

func (dnc *dataNodeClient) getVegaTime() (int64, error) {
	vegaTimeResp, err := dnc.datanode.GetVegaTime(context.Background(), &dn.GetVegaTimeRequest{})
	if err != nil {
		return 0, err
	}
	return vegaTimeResp.Timestamp, nil
}

func (dnc *dataNodeClient) listDelegations() ([]*vega.Delegation, error) {
	delegationResp, err := dnc.datanode.ListDelegations(context.Background(), &dn.ListDelegationsRequest{})
	if err != nil {
		return nil, err
	}
	delegations := make([]*vega.Delegation, 0, len(delegationResp.Delegations.Edges))
	for _, d := range delegationResp.Delegations.Edges {
		delegations = append(delegations, d.Node)
	}
	return delegations, nil
}

func (dnc *dataNodeClient) getEpoch() (*vega.Epoch, error) {
	epochResp, err := dnc.datanode.GetEpoch(context.Background(), &dn.GetEpochRequest{})
	if err != nil {
		return nil, err
	}

	return &vega.Epoch{
		Seq:        epochResp.Epoch.Seq,
		Timestamps: epochResp.Epoch.Timestamps,
	}, nil
}

func (dnc *dataNodeClient) listNodes() ([]*vega.Node, error) {
	nodeResp, err := dnc.datanode.ListNodes(context.Background(), &dn.ListNodesRequest{})
	if err != nil {
		return nil, err
	}
	nodes := make([]*vega.Node, 0, len(nodeResp.Nodes.Edges))
	for _, ne := range nodeResp.Nodes.Edges {
		nodes = append(nodes, &vega.Node{
			Id:              ne.Node.Id,
			PubKey:          ne.Node.PubKey,
			TmPubKey:        ne.Node.TmPubKey,
			EthereumAddress: crypto.EthereumChecksumAddress(ne.Node.EthereumAddress),
			InfoUrl:         ne.Node.InfoUrl,
			Location:        ne.Node.Location,
			Status:          ne.Node.Status,
			RankingScore:    ne.Node.RankingScore,
			Name:            ne.Node.Name,
			AvatarUrl:       ne.Node.AvatarUrl,
		})
	}
	return nodes, nil
}

func (dnc *dataNodeClient) listNetworkParameters() ([]*vega.NetworkParameter, error) {
	resp, err := dnc.datanode.ListNetworkParameters(context.Background(), &dn.ListNetworkParametersRequest{})
	if err != nil {
		return nil, err
	}

	params := make([]*vega.NetworkParameter, 0, len(resp.NetworkParameters.Edges))
	for _, npe := range resp.NetworkParameters.Edges {
		params = append(params, npe.Node)
	}
	return params, nil
}

func (dnc *dataNodeClient) listGovernanceData() ([]*vega.Proposal, error) {
	resp, err := dnc.datanode.ListGovernanceData(context.Background(), &dn.ListGovernanceDataRequest{})
	if err != nil {
		return nil, err
	}
	proposals := make([]*vega.Proposal, 0, len(resp.Connection.Edges))
	for _, gde := range resp.Connection.Edges {
		if gde.Node.Proposal.State != vega.Proposal_STATE_DECLINED && gde.Node.Proposal.State != vega.Proposal_STATE_REJECTED {
			proposals = append(proposals, gde.Node.Proposal)
		}
	}
	return proposals, nil
}

func (dnc *dataNodeClient) listDeposits() ([]*vega.Deposit, error) {
	resp, err := dnc.datanode.ListDeposits(context.Background(), &dn.ListDepositsRequest{})
	if err != nil {
		return nil, err
	}
	deposits := make([]*vega.Deposit, 0, len(resp.Deposits.Edges))
	for _, de := range resp.Deposits.Edges {
		deposits = append(deposits, de.Node)
	}
	return deposits, nil
}

func (dnc *dataNodeClient) listWithdrawals() ([]*vega.Withdrawal, error) {
	resp, err := dnc.datanode.ListWithdrawals(context.Background(), &dn.ListWithdrawalsRequest{})
	if err != nil {
		return nil, err
	}
	withdrawals := make([]*vega.Withdrawal, 0, len(resp.Withdrawals.Edges))
	for _, we := range resp.Withdrawals.Edges {
		we.Node.Ext = nil
		withdrawals = append(withdrawals, we.Node)
	}
	return withdrawals, nil
}

func (dnc *dataNodeClient) listTransfers() ([]*v1.Transfer, error) {
	resp, err := dnc.datanode.ListTransfers(context.Background(), &dn.ListTransfersRequest{})
	if err != nil {
		return nil, err
	}
	transfers := make([]*v1.Transfer, 0, len(resp.Transfers.Edges))
	for _, te := range resp.Transfers.Edges {
		transfers = append(transfers, te.Node)
	}
	return transfers, nil
}

func (dnc *dataNodeClient) listPositions() ([]*vega.Position, error) {
	resp, err := dnc.datanode.ListPositions(context.Background(), &dn.ListPositionsRequest{})
	if err != nil {
		return nil, err
	}
	positions := make([]*vega.Position, 0, len(resp.Positions.Edges))
	for _, pe := range resp.Positions.Edges {
		positions = append(positions, pe.Node)
	}
	return positions, nil
}

func (dnc *dataNodeClient) listLiquidityProvisions(market string) ([]*vega.LiquidityProvision, error) {
	live := true
	resp, err := dnc.datanode.ListLiquidityProvisions(context.Background(), &dn.ListLiquidityProvisionsRequest{MarketId: &market, Live: &live})
	if err != nil {
		return nil, err
	}
	lps := make([]*vega.LiquidityProvision, 0, len(resp.LiquidityProvisions.Edges))
	for _, lpe := range resp.LiquidityProvisions.Edges {
		if lpe.Node.Status == vega.LiquidityProvision_STATUS_PENDING ||  lpe.Node.Status == vega.LiquidityProvision_STATUS_ACTIVE || lpe.Node.Status == vega.LiquidityProvision_STATUS_UNDEPLOYED {
			lps = append(lps, lpe.Node)
		}
	}
	return lps, nil
}

func (dnc *dataNodeClient) getStake() ([]*v1.StakeLinking, error) {
	parties, err := dnc.listParties()
	if err != nil {
		return nil, err
	}

	stake := []*v1.StakeLinking{}
	for _, p := range parties {
		resp, err := dnc.datanode.GetStake(context.Background(), &dn.GetStakeRequest{PartyId: p.Id})
		if err != nil {
			return stake, err
		}
		for _, sle := range resp.StakeLinkings.Edges {
			// ignore 0 amounts
			if sle.Node.Amount != "0" {
				stake = append(stake, sle.Node)
			}
		}
	}
	return stake, nil
}
