package checkpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code.vegaprotocol.io/protos/vega"
	events "code.vegaprotocol.io/protos/vega/events/v1"
	snapshot "code.vegaprotocol.io/protos/vega/snapshot/v1"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

type all struct {
	Governance *snapshot.Proposals  `json:"governance_proposals,omitempty"`
	Assets     *snapshot.Assets     `json:"assets,omitempty"`
	Collateral *snapshot.Collateral `json:"collateral,omitempty"`
	NetParams  *snapshot.NetParams  `json:"network_parameters,omitempty"`
	Delegate   *snapshot.Delegate   `json:"delegate,omitempty"`
	Epoch      *events.EpochEvent   `json:"epoch,omitempty"`
	Block      *snapshot.Block      `json:"block,omitempty"`
}

// AssetErr a convenience error type
type AssetErr []error

func (a all) CheckAssetsCollateral() error {
	assets := make(map[string]struct{}, len(a.Assets.Assets))
	for _, e := range a.Assets.Assets {
		assets[e.Id] = struct{}{}
	}
	cAssets := make(map[string]struct{}, len(assets)) // should be no more than total assets
	for _, c := range a.Collateral.Balances {
		cAssets[c.Asset] = struct{}{}
	}
	errs := []error{}
	for a := range cAssets {
		if _, ok := assets[a]; !ok {
			errs = append(errs, fmt.Errorf("collateral contains '%s' asset, asset checkpoint does not", a))
		}
	}
	if len(errs) != 0 {
		return AssetErr(errs)
	}
	return nil
}

func (a all) JSON() ([]byte, error) {
	// format nicely
	marshaler := jsonpb.Marshaler{
		Indent: "   ",
	}
	g, err := marshaler.MarshalToString(a.Governance)
	if err != nil {
		return nil, err
	}
	as, err := marshaler.MarshalToString(a.Assets)
	if err != nil {
		return nil, err
	}
	c, err := marshaler.MarshalToString(a.Collateral)
	if err != nil {
		return nil, err
	}
	n, err := marshaler.MarshalToString(a.NetParams)
	if err != nil {
		return nil, err
	}
	d, err := marshaler.MarshalToString(a.Delegate)
	if err != nil {
		return nil, err
	}
	e, err := marshaler.MarshalToString(a.Epoch)
	if err != nil {
		return nil, err
	}
	all := allJSON{
		Governance: json.RawMessage(g),
		Assets:     json.RawMessage(as),
		Collateral: json.RawMessage(c),
		NetParams:  json.RawMessage(n),
		Delegate:   json.RawMessage(d),
		Epoch:      json.RawMessage(e),
	}
	b, err := json.MarshalIndent(all, "", "   ")
	if err != nil {
		return nil, err
	}
	return b, nil
}

// FromJSON can be used in the future to load JSON input and generate a checkpoint file
func (a *all) FromJSON(in []byte) error {
	all := &allJSON{}
	if err := json.Unmarshal(in, all); err != nil {
		return err
	}
	if len(all.Governance) != 0 {
		a.Governance = &snapshot.Proposals{}
		reader := bytes.NewReader([]byte(all.Governance))
		if err := jsonpb.Unmarshal(reader, a.Governance); err != nil {
			return err
		}
	}
	if len(all.Assets) != 0 {
		a.Assets = &snapshot.Assets{}
		reader := bytes.NewReader([]byte(all.Assets))
		if err := jsonpb.Unmarshal(reader, a.Assets); err != nil {
			return err
		}
	}
	if len(all.Collateral) != 0 {
		a.Collateral = &snapshot.Collateral{}
		reader := bytes.NewReader([]byte(all.Collateral))
		if err := jsonpb.Unmarshal(reader, a.Collateral); err != nil {
			return err
		}
	}
	if len(all.NetParams) != 0 {
		a.NetParams = &snapshot.NetParams{}
		reader := bytes.NewReader([]byte(all.NetParams))
		if err := jsonpb.Unmarshal(reader, a.NetParams); err != nil {
			return err
		}
	}
	if len(all.Delegate) != 0 {
		a.Delegate = &snapshot.Delegate{}
		reader := bytes.NewReader([]byte(all.Delegate))
		if err := jsonpb.Unmarshal(reader, a.Delegate); err != nil {
			return err
		}
	}
	if len(all.Epoch) != 0 {
		a.Epoch = &events.EpochEvent{}
		reader := bytes.NewReader([]byte(all.Epoch))
		if err := jsonpb.Unmarshal(reader, a.Epoch); err != nil {
			return err
		}
	}
	if len(all.Block) != 0 {
		a.Block = &snapshot.Block{}
		reader := bytes.NewReader([]byte(all.Block))
		if err := jsonpb.Unmarshal(reader, a.Block); err != nil {
			return err
		}
	}

	return nil
}

// Hash returns the hash for a checkpoint (copied form core repo - needs to be kept in sync)
func Hash(data []byte) []byte {
	h := sha3.New256()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func (a all) SnapshotData() ([]byte, error) {
	g, err := proto.Marshal(a.Governance)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(a.Collateral)
	if err != nil {
		return nil, err
	}
	n, err := proto.Marshal(a.NetParams)
	if err != nil {
		return nil, err
	}
	d, err := proto.Marshal(a.Delegate)
	if err != nil {
		return nil, err
	}
	e, err := proto.Marshal(a.Epoch)
	if err != nil {
		return nil, err
	}
	b, err := proto.Marshal(a.Block)
	if err != nil {
		return nil, err
	}
	cp := &snapshot.Checkpoint{
		Governance:        g,
		Collateral:        c,
		NetworkParameters: n,
		Delegation:        d,
		Epoch:             e,
		Block:             b,
	}
	if cp.Assets, err = proto.Marshal(a.Assets); err != nil {
		return nil, err
	}
	ret, err := proto.Marshal(cp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Error outputs the mismatches in an easy to read way
func (a AssetErr) Error() string {
	out := make([]string, 0, len(a)+1)
	out = append(out, "unexpected asset/collateral data found:")
	for _, e := range a {
		out = append(out, fmt.Sprintf("\t%s", e.Error()))
	}
	return strings.Join(out, "\n")
}

func dummy() *all {
	ae := &snapshot.AssetEntry{
		Id: "ETH",
		AssetDetails: &vega.AssetDetails{
			Name:        "ETH",
			Symbol:      "ETH",
			TotalSupply: "100000000000",
			Decimals:    5,
			MinLpStake:  "",
			Source: &vega.AssetDetails_BuiltinAsset{
				BuiltinAsset: &vega.BuiltinAsset{
					MaxFaucetAmountMint: "100000000000",
				},
			},
		},
	}
	bal := &snapshot.AssetBalance{
		Party:   "deadbeef007",
		Asset:   "ETH",
		Balance: "1000000",
	}
	prop := &vega.Proposal{
		Id:        "prop-1",
		Reference: "dummy-proposal",
		PartyId:   "deadbeef007",
		State:     vega.Proposal_STATE_ENACTED,
		Timestamp: time.Now().Add(-1 * time.Hour).Unix(),
		Terms: &vega.ProposalTerms{
			ClosingTimestamp:    time.Now().Add(24 * time.Hour).Unix(),
			EnactmentTimestamp:  time.Now().Add(-10 * time.Minute).Unix(),
			ValidationTimestamp: time.Now().Add(-1*time.Hour - time.Second).Unix(),
			Change: &vega.ProposalTerms_NewMarket{
				NewMarket: &vega.NewMarket{
					Changes: &vega.NewMarketConfiguration{
						Instrument: &vega.InstrumentConfiguration{
							Name: "ETH/FOO",
							Code: "bar",
							Product: &vega.InstrumentConfiguration_Future{
								Future: &vega.FutureProduct{ // omitted oracle spec for now
									Maturity:        time.Now().Add(48 * time.Hour).Format(time.RFC3339),
									SettlementAsset: "ETH",
									QuoteName:       "ETH",
								},
							},
						},
						DecimalPlaces: 5,
						PriceMonitoringParameters: &vega.PriceMonitoringParameters{
							Triggers: []*vega.PriceMonitoringTrigger{
								{
									Horizon:          10,
									Probability:      .95,
									AuctionExtension: 10,
								},
							},
						},
						LiquidityMonitoringParameters: &vega.LiquidityMonitoringParameters{
							TargetStakeParameters: &vega.TargetStakeParameters{
								TimeWindow:    10,
								ScalingFactor: 0.7,
							},
							TriggeringRatio:  0.5,
							AuctionExtension: 10,
						},
						RiskParameters: &vega.NewMarketConfiguration_LogNormal{
							LogNormal: &vega.LogNormalRiskModel{
								RiskAversionParameter: 0.1,
								Tau:                   0.2,
								Params: &vega.LogNormalModelParams{
									Mu:    0.3,
									R:     0.3,
									Sigma: 0.3,
								},
							},
						},
						TradingMode: &vega.NewMarketConfiguration_Continuous{
							Continuous: &vega.ContinuousTrading{
								TickSize: "1",
							},
						},
					},
					LiquidityCommitment: &vega.NewMarketCommitment{
						CommitmentAmount: "100",
						Fee:              ".0001",
						Sells: []*vega.LiquidityOrder{
							{
								Reference:  vega.PeggedReference_PEGGED_REFERENCE_BEST_BID,
								Proportion: 10,
								Offset:     -5,
							},
						},
						Buys: []*vega.LiquidityOrder{
							{
								Reference:  vega.PeggedReference_PEGGED_REFERENCE_MID,
								Proportion: 10,
								Offset:     5,
							},
						},
						Reference: "dummy-LP",
					},
				},
			},
		},
	}
	del := &snapshot.Delegate{
		Active: []*snapshot.DelegateEntry{
			{
				Party:    "deadbeef007",
				Node:     "node0",
				Amount:   "100",
				EpochSeq: 0,
			},
		},
		Pending: []*snapshot.DelegateEntry{
			{
				Party:      "deadbeef007",
				Node:       "node0",
				Amount:     "100",
				Undelegate: true,
				EpochSeq:   1,
			},
		},
	}
	t := time.Now()
	return &all{
		Assets: &snapshot.Assets{
			Assets: []*snapshot.AssetEntry{ae},
		},
		Collateral: &snapshot.Collateral{
			Balances: []*snapshot.AssetBalance{bal},
		},
		Governance: &snapshot.Proposals{
			Proposals: []*vega.Proposal{prop},
		},
		NetParams: &snapshot.NetParams{
			Params: []*vega.NetworkParameter{
				{
					Key:   "foo",
					Value: "bar",
				},
			},
		},
		Delegate: del,
		Epoch: &events.EpochEvent{
			Seq:        0,
			Action:     vega.EpochAction_EPOCH_ACTION_START,
			StartTime:  t.UnixNano(),
			ExpireTime: t.Add(24 * time.Hour).UnixNano(),
			EndTime:    t.Add(25 * time.Hour).UnixNano(),
		},
		Block: &snapshot.Block{
			Height: 1,
		},
	}
}

type allJSON struct {
	Governance json.RawMessage `json:"governance_proposals,omitempty"`
	Assets     json.RawMessage `json:"assets,omitempty"`
	Collateral json.RawMessage `json:"collateral,omitempty"`
	NetParams  json.RawMessage `json:"network_parameters,omitempty"`
	Delegate   json.RawMessage `json:"delegate,omitempty"`
	Epoch      json.RawMessage `json:"epoch,omitempty"`
	Block      json.RawMessage `json:"block,omitempty"`
}
