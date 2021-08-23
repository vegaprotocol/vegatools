package checkpoint

import (
	"encoding/json"

	"code.vegaprotocol.io/protos/vega"
	snapshot "code.vegaprotocol.io/protos/vega/snapshot/v1"
)

type ADetails struct {
	*vega.AssetDetails
	Source json.RawMessage `json:"Source"`
}

type Asset struct {
	Id      string    `json:"id"`
	Details *ADetails `json:"asset_details"`
}

type PropTerms struct {
	*vega.ProposalTerms
	Change json.RawMessage `json:"Change"`
}

type Prop struct {
	*vega.Proposal
	Terms *PropTerms `json:"terms"`
}

type Assets struct {
	Assets []*Asset `json:"assets,omitempty"`
}

type Proposals struct {
	Proposals []*Prop `json:"proposals,omitempty"`
}

type JSONCheckpoint struct {
	Governance *Proposals           `json:"governance_proposals,omitempty"`
	Assets     *Assets              `json:"assets,omitempty"`
	Collateral *snapshot.Collateral `json:"collateral,omitempty"`
	NetParams  *snapshot.NetParams  `json:"network_parameters,omitempty"`
}

func (j *JSONCheckpoint) FromJSON(data []byte) error {
	return json.Unmarshal(data, j)
}

// ToAll will take care of all the one-of fields by calling the correct methods
// before assigning to the all type fields
func (j *JSONCheckpoint) ToAll() *all {
	r := &all{
		Collateral: j.Collateral,
		NetParams:  j.NetParams,
		Governance: j.Governance.ToSnapshot(),
		Assets:     j.Assets.ToSnapshot(),
	}
	return r
}

func (a *Assets) ToSnapshot() *snapshot.Assets {
	ret := &snapshot.Assets{
		Assets: make([]*snapshot.AssetEntry, 0, len(a.Assets)),
	}
	for _, as := range a.Assets {
		ret.Assets = append(ret.Assets, &snapshot.AssetEntry{
			Id:           as.Id,
			AssetDetails: as.Details.GetDetails(),
		})
	}
	return ret
}

func (a *ADetails) UnmarshalSource() error {
	// try various types and set the correct one after unmarshalling
	// ERC20?
	src := &vega.AssetDetails_Erc20{}
	if err := json.Unmarshal([]byte(a.Source), src); err == nil {
		a.AssetDetails.Source = src
		return nil
	}
	builtIn := &vega.AssetDetails_BuiltinAsset{}
	if err := json.Unmarshal([]byte(a.Source), builtIn); err != nil {
		return err
	}
	a.AssetDetails.Source = builtIn
	return nil
}

func (a ADetails) GetDetails() *vega.AssetDetails {
	if a.AssetDetails.Source == nil {
		_ = a.UnmarshalSource()
	}
	return a.AssetDetails
}

func (p Proposals) ToSnapshot() *snapshot.Proposals {
	ret := &snapshot.Proposals{
		Proposals: make([]*vega.Proposal, 0, len(p.Proposals)),
	}
	for _, pr := range p.Proposals {
		// convert
		pr.Proposal.Terms = pr.Terms.GetTerms()
		ret.Proposals = append(ret.Proposals, pr.Proposal)
	}
	return ret
}

type NewAsset struct {
	*vega.NewAsset
	Changes *ADetails
}

type NewMarket struct {
	*vega.ProposalTerms_NewMarket
	NewMarket *PT_NM `json:"NewMarket"`
}

type PT_NM struct {
	*vega.NewMarket
	Changes *NM_Changes `json:"Changes"`
}

type NM_Changes struct {
	*vega.NewMarketConfiguration
	Instrument     *NM_Instrument  `json:"instrument"`
	RiskParameters json.RawMessage `json:"RiskParameters"`
	TradingMode    json.RawMessage `json:"TradingMode"`
}

type NM_Instrument struct {
	*vega.InstrumentConfiguration
	Product *vega.InstrumentConfiguration_Future `json:"Product,omitempty"`
}

func (p *PropTerms) UnmarshalTerms() error {
	um := &vega.ProposalTerms_UpdateMarket{}
	if err := json.Unmarshal([]byte(p.Change), um); err == nil {
		p.ProposalTerms.Change = um
		return nil
	}
	nm := &NewMarket{}
	if err := json.Unmarshal([]byte(p.Change), nm); err == nil {
		p.ProposalTerms.Change = nm.GetPTNM()
		return nil
	}
	un := &vega.ProposalTerms_UpdateNetworkParameter{}
	if err := json.Unmarshal([]byte(p.Change), un); err == nil {
		p.ProposalTerms.Change = un
		return nil
	}
	na := &NewAsset{}
	if err := json.Unmarshal([]byte(p.Change), na); err != nil {
		return err
	}
	na.NewAsset.Changes = na.Changes.GetDetails()
	p.ProposalTerms.Change = &vega.ProposalTerms_NewAsset{
		NewAsset: na.NewAsset,
	}
	return nil
}

func (p PropTerms) GetTerms() *vega.ProposalTerms {
	if p.ProposalTerms.Change == nil {
		_ = p.UnmarshalTerms()
	}
	return p.ProposalTerms
}

func (n *NewMarket) GetPTNM() *vega.ProposalTerms_NewMarket {
	// populate instrument
	n.NewMarket.Changes.Instrument.InstrumentConfiguration.Product = n.NewMarket.Changes.Instrument.Product
	// Types that are valid to be assigned to RiskParameters:
	//	*NewMarketConfiguration_Simple
	//	*NewMarketConfiguration_LogNormal
	// RiskParameters isNewMarketConfiguration_RiskParameters `protobuf_oneof:"risk_parameters"`
	// Trading mode for the new market
	//
	// Types that are valid to be assigned to TradingMode:
	//	*NewMarketConfiguration_Continuous
	//	*NewMarketConfiguration_Discrete
	riskParams := []byte(n.NewMarket.Changes.RiskParameters)
	rs := &vega.NewMarketConfiguration_Simple{}
	if err := json.Unmarshal(riskParams, rs); err != nil {
		rl := &vega.NewMarketConfiguration_LogNormal{}
		if err := json.Unmarshal(riskParams, rl); err != nil {
			return nil
		}
		n.NewMarket.Changes.NewMarketConfiguration.RiskParameters = rl
	} else {
		n.NewMarket.Changes.NewMarketConfiguration.RiskParameters = rs
	}
	trading := []byte(n.NewMarket.Changes.TradingMode)
	tc := &vega.NewMarketConfiguration_Continuous{}
	if err := json.Unmarshal(trading, tc); err != nil {
		td := &vega.NewMarketConfiguration_Discrete{}
		if err := json.Unmarshal(trading, td); err != nil {
			return nil
		}
		n.NewMarket.Changes.NewMarketConfiguration.TradingMode = td
	} else {
		n.NewMarket.Changes.NewMarketConfiguration.TradingMode = tc
	}
	n.NewMarket.NewMarket.Changes = n.NewMarket.Changes.NewMarketConfiguration
	n.ProposalTerms_NewMarket.NewMarket = n.NewMarket.NewMarket
	return n.ProposalTerms_NewMarket
}
