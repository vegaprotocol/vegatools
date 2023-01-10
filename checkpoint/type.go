package checkpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code.vegaprotocol.io/vega/protos/vega"
	checkpoint "code.vegaprotocol.io/vega/protos/vega/checkpoint/v1"
	events "code.vegaprotocol.io/vega/protos/vega/events/v1"

	"github.com/gogo/protobuf/jsonpb"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

type all struct {
	Governance      *checkpoint.Proposals       `json:"governance_proposals,omitempty"`
	Assets          *checkpoint.Assets          `json:"assets,omitempty"`
	Collateral      *checkpoint.Collateral      `json:"collateral,omitempty"`
	NetParams       *checkpoint.NetParams       `json:"network_parameters,omitempty"`
	Delegate        *checkpoint.Delegate        `json:"delegate,omitempty"`
	Epoch           *events.EpochEvent          `json:"epoch,omitempty"`
	Block           *checkpoint.Block           `json:"block,omitempty"`
	Rewards         *checkpoint.Rewards         `json:"rewards,omitempty"`
	Banking         *checkpoint.Banking         `json:"banking,omitempty"`
	Validators      *checkpoint.Validators      `json:"validators,omitempty"`
	Staking         *checkpoint.Staking         `json:"staking,omitempty"`
	MultisigControl *checkpoint.MultisigControl `json:"multisig_control,omitempty"`
	MarketTracker   *checkpoint.MarketTracker   `json:"market_tracker,omitempty"`
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
	r, err := marshaler.MarshalToString(a.Rewards)
	if err != nil {
		return nil, err
	}

	block, err := marshaler.MarshalToString(a.Block)
	if err != nil {
		return nil, err
	}
	banking, err := marshaler.MarshalToString(a.Banking)
	if err != nil {
		return nil, err
	}

	validators, err := marshaler.MarshalToString(a.Validators)
	if err != nil {
		return nil, err
	}

	staking, err := marshaler.MarshalToString(a.Staking)
	if err != nil {
		return nil, err
	}

	multisig, err := marshaler.MarshalToString(a.MultisigControl)
	if err != nil {
		return nil, err
	}

	marketTracker, err := marshaler.MarshalToString(a.MarketTracker)
	if err != nil {
		return nil, err
	}

	all := allJSON{
		Governance:      json.RawMessage(g),
		Assets:          json.RawMessage(as),
		Collateral:      json.RawMessage(c),
		NetParams:       json.RawMessage(n),
		Delegate:        json.RawMessage(d),
		Epoch:           json.RawMessage(e),
		Block:           json.RawMessage(block),
		Rewards:         json.RawMessage(r),
		Banking:         json.RawMessage(banking),
		Validators:      json.RawMessage(validators),
		Staking:         json.RawMessage(staking),
		MultisigControl: json.RawMessage(multisig),
		MarketTracker:   json.RawMessage(marketTracker),
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
		a.Governance = &checkpoint.Proposals{}
		reader := bytes.NewReader([]byte(all.Governance))
		if err := jsonpb.Unmarshal(reader, a.Governance); err != nil {
			return err
		}
	}
	if len(all.Assets) != 0 {
		a.Assets = &checkpoint.Assets{}
		reader := bytes.NewReader([]byte(all.Assets))
		if err := jsonpb.Unmarshal(reader, a.Assets); err != nil {
			return err
		}
	}
	if len(all.Collateral) != 0 {
		a.Collateral = &checkpoint.Collateral{}
		reader := bytes.NewReader([]byte(all.Collateral))
		if err := jsonpb.Unmarshal(reader, a.Collateral); err != nil {
			return err
		}
	}
	if len(all.NetParams) != 0 {
		a.NetParams = &checkpoint.NetParams{}
		reader := bytes.NewReader([]byte(all.NetParams))
		if err := jsonpb.Unmarshal(reader, a.NetParams); err != nil {
			return err
		}
	}
	if len(all.Delegate) != 0 {
		a.Delegate = &checkpoint.Delegate{}
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
		a.Block = &checkpoint.Block{}
		reader := bytes.NewReader([]byte(all.Block))
		if err := jsonpb.Unmarshal(reader, a.Block); err != nil {
			return err
		}
	}
	if len(all.Rewards) != 0 {
		a.Rewards = &checkpoint.Rewards{}
		reader := bytes.NewReader([]byte(all.Rewards))
		if err := jsonpb.Unmarshal(reader, a.Rewards); err != nil {
			return err
		}
	}

	if len(all.Banking) != 0 {
		a.Banking = &checkpoint.Banking{}
		reader := bytes.NewReader([]byte(all.Banking))
		if err := jsonpb.Unmarshal(reader, a.Banking); err != nil {
			return err
		}
	}

	if len(all.Validators) != 0 {
		a.Validators = &checkpoint.Validators{}
		reader := bytes.NewReader([]byte(all.Validators))
		if err := jsonpb.Unmarshal(reader, a.Validators); err != nil {
			return err
		}
	}

	if len(all.Staking) != 0 {
		a.Staking = &checkpoint.Staking{}
		reader := bytes.NewReader([]byte(all.Staking))
		if err := jsonpb.Unmarshal(reader, a.Staking); err != nil {
			return err
		}
	}
	if len(all.MultisigControl) != 0 {
		a.MultisigControl = &checkpoint.MultisigControl{}
		reader := bytes.NewReader([]byte(all.MultisigControl))
		if err := jsonpb.Unmarshal(reader, a.MultisigControl); err != nil {
			return err
		}
	}

	if len(all.MarketTracker) != 0 {
		a.MarketTracker = &checkpoint.MarketTracker{}
		reader := bytes.NewReader([]byte(all.MarketTracker))
		if err := jsonpb.Unmarshal(reader, a.MarketTracker); err != nil {
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

func hashBytes(cp *checkpoint.Checkpoint) []byte {
	ret := make([]byte, 0, len(cp.Governance)+len(cp.Assets)+len(cp.Collateral)+len(cp.NetworkParameters)+len(cp.Delegation)+len(cp.Epoch)+len(cp.Block)+len(cp.Rewards)+len(cp.Banking)+len(cp.Validators)+len(cp.Staking)+len(cp.MultisigControl))
	// the order in which we append is quite important
	ret = append(ret, cp.NetworkParameters...)
	ret = append(ret, cp.Assets...)
	ret = append(ret, cp.Collateral...)
	ret = append(ret, cp.Delegation...)
	ret = append(ret, cp.Epoch...)
	ret = append(ret, cp.Block...)
	ret = append(ret, cp.Governance...)
	ret = append(ret, cp.Rewards...)
	ret = append(ret, cp.Banking...)
	ret = append(ret, cp.Validators...)
	ret = append(ret, cp.Staking...)
	return append(ret, cp.MultisigControl...)
}

func (a all) CheckpointData() ([]byte, []byte, error) {
	g, err := proto.Marshal(a.Governance)
	if err != nil {
		return nil, nil, err
	}
	c, err := proto.Marshal(a.Collateral)
	if err != nil {
		return nil, nil, err
	}
	n, err := proto.Marshal(a.NetParams)
	if err != nil {
		return nil, nil, err
	}
	d, err := proto.Marshal(a.Delegate)
	if err != nil {
		return nil, nil, err
	}
	e, err := proto.Marshal(a.Epoch)
	if err != nil {
		return nil, nil, err
	}
	b, err := proto.Marshal(a.Block)
	if err != nil {
		return nil, nil, err
	}
	r, err := proto.Marshal(a.Rewards)
	if err != nil {
		return nil, nil, err
	}
	banking, err := proto.Marshal(a.Banking)
	if err != nil {
		return nil, nil, err
	}
	validators, err := proto.Marshal(a.Validators)
	if err != nil {
		return nil, nil, err
	}
	staking, err := proto.Marshal(a.Staking)
	if err != nil {
		return nil, nil, err
	}
	multi, err := proto.Marshal(a.MultisigControl)
	if err != nil {
		return nil, nil, err
	}
	marketTracker, err := proto.Marshal(a.MarketTracker)
	if err != nil {
		return nil, nil, err
	}
	cp := &checkpoint.Checkpoint{
		Governance:        g,
		Collateral:        c,
		NetworkParameters: n,
		Delegation:        d,
		Epoch:             e,
		Block:             b,
		Rewards:           r,
		Banking:           banking,
		Validators:        validators,
		Staking:           staking,
		MultisigControl:   multi,
		MarketTracker:     marketTracker,
	}
	if cp.Assets, err = proto.Marshal(a.Assets); err != nil {
		return nil, nil, err
	}
	ret, err := proto.Marshal(cp)
	if err != nil {
		return nil, nil, err
	}
	hb := hashBytes(cp)
	return ret, hb, nil
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
	ae := &checkpoint.AssetEntry{
		Id: "ETH",
		AssetDetails: &vega.AssetDetails{
			Name:     "ETH",
			Symbol:   "ETH",
			Decimals: 5,
			Quantum:  "",
			Source: &vega.AssetDetails_BuiltinAsset{
				BuiltinAsset: &vega.BuiltinAsset{
					MaxFaucetAmountMint: "100000000000",
				},
			},
		},
	}
	bal := &checkpoint.AssetBalance{
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
									Probability:      "0.95",
									AuctionExtension: 10,
								},
							},
						},
						LiquidityMonitoringParameters: &vega.LiquidityMonitoringParameters{
							TargetStakeParameters: &vega.TargetStakeParameters{
								TimeWindow:    10,
								ScalingFactor: 0.7,
							},
							TriggeringRatio:  "0.5",
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
					},
				},
			},
		},
	}
	del := &checkpoint.Delegate{
		Active: []*checkpoint.DelegateEntry{
			{
				Party:    "deadbeef007",
				Node:     "node0",
				Amount:   "100",
				EpochSeq: 0,
			},
		},
		Pending: []*checkpoint.DelegateEntry{
			{
				Party:      "deadbeef007",
				Node:       "node0",
				Amount:     "100",
				Undelegate: true,
				EpochSeq:   1,
			},
		},
		AutoDelegation: []string{
			"deadbeef007",
		},
	}
	t := time.Now()
	return &all{
		Assets: &checkpoint.Assets{
			Assets: []*checkpoint.AssetEntry{ae},
		},
		Collateral: &checkpoint.Collateral{
			Balances: []*checkpoint.AssetBalance{bal},
		},
		Governance: &checkpoint.Proposals{
			Proposals: []*vega.Proposal{prop},
		},
		NetParams: &checkpoint.NetParams{
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
		Block: &checkpoint.Block{
			Height: 1,
		},
		Banking: &checkpoint.Banking{
			RecurringTransfers: &checkpoint.RecurringTransfers{
				RecurringTransfers: []*events.Transfer{
					{
						Id:              "someid",
						From:            "somefrom",
						FromAccountType: vega.AccountType_ACCOUNT_TYPE_GENERAL,
						To:              "someto",
						ToAccountType:   vega.AccountType_ACCOUNT_TYPE_GENERAL,
						Asset:           "someasset",
						Amount:          "100",
						Reference:       "someref",
						Status:          events.Transfer_STATUS_PENDING,
						Kind: &events.Transfer_Recurring{
							Recurring: &events.RecurringTransfer{
								StartEpoch: 10,
								EndEpoch:   toPtr(uint64(100)),
								Factor:     "1",
							},
						},
					},
				},
			},
		},
	}
}

type allJSON struct {
	Governance      json.RawMessage `json:"governance_proposals,omitempty"`
	Assets          json.RawMessage `json:"assets,omitempty"`
	Collateral      json.RawMessage `json:"collateral,omitempty"`
	NetParams       json.RawMessage `json:"network_parameters,omitempty"`
	Delegate        json.RawMessage `json:"delegate,omitempty"`
	Epoch           json.RawMessage `json:"epoch,omitempty"`
	Block           json.RawMessage `json:"block,omitempty"`
	Rewards         json.RawMessage `json:"rewards,omitempty"`
	KeyRotations    json.RawMessage `json:"key_rotations,omitempty"`
	Banking         json.RawMessage `json:"banking,omitempty"`
	Validators      json.RawMessage `json:"validators,omitempty"`
	Staking         json.RawMessage `json:"staking,omitempty"`
	MultisigControl json.RawMessage `json:"multisig_control,omitempty"`
	MarketTracker   json.RawMessage `json:"market_tracker,omitempty"`
}

func toPtr[T any](t T) *T { return &t }
