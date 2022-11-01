package diff

import (
	"fmt"

	dn "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"code.vegaprotocol.io/vega/protos/vega"
	v1 "code.vegaprotocol.io/vega/protos/vega/events/v1"
)

// MatchResult represents the result of comparing core snapshot of an engine with data node api for the same.
type MatchResult int64

const (
	// FullMatch means no discrepancies found.
	FullMatch MatchResult = iota
	// SizeMismatch means some expected entries are missing.
	SizeMismatch
	// ValuesMismatch means some values disagree between the core snapshot and the data node api.
	ValuesMismatch
)

var matchResultToName map[MatchResult]string = map[MatchResult]string{
	FullMatch:      "full match",
	SizeMismatch:   "mismatching number of elements",
	ValuesMismatch: "mismatching values",
}

// Result corresponds to a dataset representing data node state ot core snapshot state.
type Result struct {
	Accounts    []*dn.AccountBalance
	Orders      []*vega.Order
	Markets     []*vega.Market
	Parties     []*vega.Party
	Limits      *vega.NetworkLimits
	Assets      []*vega.Asset
	VegaTime    int64
	Delegations []*vega.Delegation
	Epoch       *vega.Epoch
	Nodes       []*vega.Node
	NetParams   []*vega.NetworkParameter
	Proposals   []*vega.Proposal
	Deposits    []*vega.Deposit
	Withdrawals []*vega.Withdrawal
	Transfers   []*v1.Transfer
	Positions   []*vega.Position
	Lps         []*vega.LiquidityProvision
	Stake       []*v1.StakeLinking
}

// Status is a diff summary report for a key.
type Status struct {
	Key         string
	MatchResult MatchResult
	DatanodeRes string
	CoreRes     string
	CoreResLen  int
	DataNodeLen int
}

func (ds Status) String() string {
	return fmt.Sprintf("key=%s, matchResult=%s, coreLength=%d, datanodeLength=%d, coreResult=%s, datanodeResult=%s", ds.Key, matchResultToName[ds.MatchResult], ds.CoreResLen, ds.DataNodeLen, ds.CoreRes, ds.DatanodeRes)
}

// Report is the top level diff result aggregating the results from all compared keys.
type Report struct {
	coreResult     *Result
	datanodeResult *Result
	DiffResult     []Status
	Success        bool
}

func (dr *Report) String() string {
	str := ""
	for _, ds := range dr.DiffResult {
		str += ds.String() + "\n"
	}
	return fmt.Sprintf("success=%t, report:\n%s", dr.Success, str)
}
