package diff

import (
	"io/ioutil"
	"os"

	"code.vegaprotocol.io/vega/libs/crypto"
	dn "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"code.vegaprotocol.io/vega/protos/vega"
	events "code.vegaprotocol.io/vega/protos/vega/events/v1"
	v1 "code.vegaprotocol.io/vega/protos/vega/events/v1"
	snapshot "code.vegaprotocol.io/vega/protos/vega/snapshot/v1"
	decimal "github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

type snap struct {
	chunk *snapshot.Chunk
}

// Collect returns a dataset for comparison from core snapshot.
func (s *snap) Collect() *Result {
	return &Result{
		Accounts:    s.getAccounts(),
		Orders:      s.getOrders(),
		Markets:     s.getMarkets(),
		Parties:     s.getParties(),
		Limits:      s.getNetLimits(),
		Assets:      s.getAssets(),
		VegaTime:    s.getVegaTime(),
		Delegations: s.getDelegations(),
		Epoch:       s.getEpoch(),
		Nodes:       s.getValidators(),
		NetParams:   s.getNetParams(),
		Proposals:   s.getProposals(),
		Deposits:    s.getDeposits(),
		Withdrawals: s.getWithdrawals(),
		Transfers:   s.getTransfers(),
		Positions:   s.getPositions(),
		Lps:         s.getLps(),
		Stake:       s.getStake(),
	}
}

// getNetParams returns the network parmeters from the core snapshot.
func (s *snap) getNetParams() []*vega.NetworkParameter {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_NetworkParameters:
			return c.GetNetworkParameters().Params
		default:
			continue
		}
	}
	return []*vega.NetworkParameter{}
}

// getWithdrawals returns withdrawals from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getWithdrawals() []*vega.Withdrawal {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_BankingWithdrawals:
			withdrawalsSnap := c.GetBankingWithdrawals().Withdrawals
			withdrawals := make([]*vega.Withdrawal, 0, len(withdrawalsSnap))
			for _, w := range withdrawalsSnap {
				w.Withdrawal.CreatedTimestamp = (w.Withdrawal.CreatedTimestamp / 1000) * 1000
				w.Withdrawal.WithdrawnTimestamp = (w.Withdrawal.WithdrawnTimestamp / 1000) * 1000
				w.Withdrawal.Ext = nil
				withdrawals = append(withdrawals, w.Withdrawal)
			}
			return withdrawals
		default:
			continue
		}
	}
	return []*vega.Withdrawal{}
}

// getDeposits returns deposits from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getDeposits() []*vega.Deposit {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_BankingDeposits:
			depositsSnap := c.GetBankingDeposits().Deposit
			deposits := make([]*vega.Deposit, 0, len(depositsSnap))
			for _, d := range depositsSnap {
				d.Deposit.CreatedTimestamp = (d.Deposit.CreatedTimestamp / 1000) * 1000
				d.Deposit.CreditedTimestamp = (d.Deposit.CreditedTimestamp / 1000) * 1000
				deposits = append(deposits, d.Deposit)
			}
			return deposits

		default:
			continue
		}
	}
	return []*vega.Deposit{}
}

// getLps returns the liquidity provisions from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getLps() []*vega.LiquidityProvision {
	lps := []*vega.LiquidityProvision{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_LiquidityProvisions:
			lps = append(lps, c.GetLiquidityProvisions().LiquidityProvisions...)
		default:
			continue
		}
	}
	for _, lp := range lps {
		lp.CreatedAt = (lp.CreatedAt / 1000) * 1000
		lp.UpdatedAt = (lp.UpdatedAt / 1000) * 1000
	}
	return lps
}

// getStake returns stake linking from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getStake() []*v1.StakeLinking {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_StakingAccounts:
			sl := []*v1.StakeLinking{}
			for _, sa := range c.GetStakingAccounts().Accounts {
				sl = append(sl, sa.Events...)
			}
			for _, s := range sl {
				s.FinalizedAt = (s.FinalizedAt / 1000) * 1000
			}
			return sl
		default:
			continue
		}
	}
	return []*v1.StakeLinking{}
}

// getAccounts returns account balances from the core snapshot. To make it compatible with datanode, network owner and no market are replaced with empty string.
func (s *snap) getAccounts() []*dn.AccountBalance {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_CollateralAccounts:
			accs := c.GetCollateralAccounts().Accounts
			balances := make([]*dn.AccountBalance, 0, len(accs))
			for _, a := range accs {
				owner := a.Owner
				if owner == "*" {
					owner = ""
				}
				marketID := a.MarketId
				if marketID == "!" {
					marketID = ""
				}
				balances = append(balances, &dn.AccountBalance{
					Owner:    owner,
					MarketId: marketID,
					Balance:  a.Balance,
					Asset:    a.Asset,
					Type:     a.Type,
				})
			}
			return balances

		default:
			continue
		}
	}
	return []*dn.AccountBalance{}
}

// getOrders returns the order book orders from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution. In addition price is scaled to the asset decimals to be comparable with data node.
func (s *snap) getOrders() []*vega.Order {
	orders := []*vega.Order{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_MatchingBook:
			orders = append(orders, c.GetMatchingBook().Buy...)
			orders = append(orders, c.GetMatchingBook().Sell...)
		default:
			continue
		}
	}
	assets := s.getAssets()
	markets := s.getMarkets()
	dpFactors := map[string]decimal.Decimal{}
	for _, m := range markets {
		marketDecimals := m.DecimalPlaces
		asset, _ := m.GetAsset()
		for _, a := range assets {
			if a.Id == asset {
				dpFactors[m.Id] = decimal.NewFromFloat32(10).Pow(decimal.NewFromFloat32(float32(a.Details.Decimals - marketDecimals)))
			}
		}
	}
	for _, o := range orders {
		o.CreatedAt = (o.CreatedAt / 1000) * 1000
		o.ExpiresAt = (o.ExpiresAt / 1000) * 1000
		o.UpdatedAt = (o.UpdatedAt / 1000) * 1000
		price, _ := decimal.NewFromString(o.Price)
		o.Price = price.Div(dpFactors[o.MarketId]).Truncate(0).String()
	}

	return orders
}

// getMarkets returns active markets from the core snapshot.
func (s *snap) getMarkets() []*vega.Market {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_ExecutionMarkets:
			markets := []*vega.Market{}
			for _, m := range c.GetExecutionMarkets().Markets {
				markets = append(markets, m.Market)
			}
			return markets
		default:
			continue
		}
	}
	return []*vega.Market{}
}

// getParties returns parties as a combination of parties with accounts and parties staking account. To make it comparable with datanode, network party * is replaced with "network".
func (s *snap) getParties() []*vega.Party {
	partyMap := map[string]struct{}{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_CollateralAccounts:
			for _, a := range c.GetCollateralAccounts().Accounts {
				if len(a.Owner) > 0 {
					owner := a.Owner
					if owner == "*" {
						owner = "network"
					}
					partyMap[owner] = struct{}{}
				}
			}
		case *snapshot.Payload_StakingAccounts:
			for _, a := range c.GetStakingAccounts().Accounts {
				partyMap[a.Party] = struct{}{}
			}
		default:
			continue
		}
	}
	parties := make([]*vega.Party, 0, len(partyMap))
	for k := range partyMap {
		parties = append(parties, &vega.Party{Id: k})
	}

	return parties
}

// getNetLimits returns the nework limits from the core snapshot. To work around snapshot specific logic of enabled to/from it is only set if positive.
func (s *snap) getNetLimits() *vega.NetworkLimits {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_LimitState:
			limits := c.GetLimitState()
			nl := &vega.NetworkLimits{
				CanProposeMarket:     limits.CanProposeMarket,
				CanProposeAsset:      limits.CanProposeAsset,
				GenesisLoaded:        limits.GenesisLoaded,
				ProposeMarketEnabled: limits.ProposeMarketEnabled,
				ProposeAssetEnabled:  limits.ProposeAssetEnabled,
			}
			if limits.ProposeAssetEnabledFrom > 0 {
				nl.ProposeAssetEnabledFrom = limits.ProposeAssetEnabledFrom
			}
			if limits.ProposeMarketEnabledFrom > 0 {
				nl.ProposeMarketEnabledFrom = limits.ProposeMarketEnabledFrom
			}
			return nl
		default:
			continue
		}
	}
	return &vega.NetworkLimits{}
}

// getAssets returns all pending and active assets from the core snapshot.
func (s *snap) getAssets() []*vega.Asset {
	assets := []*vega.Asset{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_ActiveAssets:
			assets = append(assets, c.GetActiveAssets().Assets...)
		case *snapshot.Payload_PendingAssets:
			assets = append(assets, c.GetPendingAssets().Assets...)
		default:
			continue
		}
	}
	return assets
}

// getVegaTime returns the vega time from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getVegaTime() int64 {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_AppState:
			return (c.GetAppState().Time / 1000) * 1000
		default:
			continue
		}
	}
	return 0
}

// getDelegations returns the delegations from the core snapshot.
func (s *snap) getDelegations() []*vega.Delegation {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_DelegationActive:
			return c.GetDelegationActive().Delegations
		default:
			continue
		}
	}
	return []*vega.Delegation{}
}

// getEpoch returns the current epoch information (timestamps)
func (s *snap) getEpoch() *vega.Epoch {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_Epoch:
			epoch := c.GetEpoch()
			return &vega.Epoch{
				Seq: epoch.Seq,
				Timestamps: &vega.EpochTimestamps{
					StartTime:  (epoch.StartTime / 1000) * 1000,
					ExpiryTime: (epoch.ExpireTime / 1000) * 1000,
				},
			}
		default:
			continue
		}
	}
	return &vega.Epoch{}
}

// getProposals returns all the pending and enacted proposals. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getProposals() []*vega.Proposal {
	pMap := map[string]*vega.Proposal{}
	proposals := []*vega.Proposal{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_GovernanceActive:
			for _, p := range c.GetGovernanceActive().Proposals {
				pMap[p.Proposal.Id] = p.Proposal
			}
		case *snapshot.Payload_GovernanceEnacted:
			for _, p := range c.GetGovernanceEnacted().Proposals {
				pMap[p.Proposal.Id] = p.Proposal
			}
		case *snapshot.Payload_GovernanceNode:
			proposals = append(proposals, c.GetGovernanceNode().Proposals...)
		default:
			continue
		}
	}
	for _, p := range pMap {
		p.Timestamp = (p.Timestamp / 1000) * 1000
		proposals = append(proposals, p)
	}
	return proposals
}

// getTransfers returns recurring and scheduled transfers from the core snapshot. To make it compatible with datanode, the timestamps are converted to have
// microsecond resolution.
func (s *snap) getTransfers() []*events.Transfer {
	transfers := []*events.Transfer{}
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_BankingRecurringTransfers:
			transfers = append(transfers, c.GetBankingRecurringTransfers().RecurringTransfers.RecurringTransfers...)
		case *snapshot.Payload_BankingScheduledTransfers:
			for _, tt := range c.GetBankingScheduledTransfers().TransfersAtTime {
				for _, t := range tt.Transfers {
					transfers = append(transfers, t.OneoffTransfer)
				}
			}
		default:
			continue
		}
	}

	for _, t := range transfers {
		t.Timestamp = (t.Timestamp / 1000) * 1000
	}
	return transfers
}

// getValidators returns information about the current validators and their ranking scores from the core snapshot. The ethereum address gets checksummed.
func (s *snap) getValidators() []*vega.Node {
	for _, c := range s.chunk.Data {
		switch c.Data.(type) {
		case *snapshot.Payload_Topology:
			nodes := []*vega.Node{}
			for _, u := range c.GetTopology().ValidatorData {
				nodes = append(nodes, &vega.Node{
					Id:              u.ValidatorUpdate.NodeId,
					PubKey:          u.ValidatorUpdate.VegaPubKey,
					TmPubKey:        u.ValidatorUpdate.TmPubKey,
					EthereumAddress: crypto.EthereumChecksumAddress(u.ValidatorUpdate.EthereumAddress),
					InfoUrl:         u.ValidatorUpdate.InfoUrl,
					Location:        u.ValidatorUpdate.Country,
					Status:          1,
					RankingScore:    u.RankingScore,
					Name:            u.ValidatorUpdate.Name,
					AvatarUrl:       u.ValidatorUpdate.AvatarUrl,
				})
			}
			return nodes
		default:
			continue
		}
	}
	return []*vega.Node{}
}

// getPositions is currently unsupported as the core snapshot and datanode have very different abstractions.
// TODO
func (s *snap) getPositions() []*vega.Position {
	return []*vega.Position{}
}

// NewSnapshotData deserealises a proto file into snap.
func newSnapshotData(fileName string) (*snap, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	chunk := snapshot.Chunk{}
	proto.Unmarshal(bytes, &chunk)

	return &snap{chunk: &chunk}, nil
}
