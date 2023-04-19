package diff

import (
	"sort"
	"strconv"

	dnproto "code.vegaprotocol.io/vega/protos/data-node/api/v2"
	"code.vegaprotocol.io/vega/protos/vega"
	v1 "code.vegaprotocol.io/vega/protos/vega/events/v1"
)

func newDiffReport(coreResult *Result, datanodeResult *Result) *Report {
	d := &Report{
		coreResult:     coreResult,
		datanodeResult: datanodeResult,
		DiffResult:     []Status{},
		Success:        true,
	}
	d.diff()
	return d
}

func (dr *Report) diff() {
	diffFuncs := []func(*Result, *Result) Status{
		diffAccountBalances,
		diffOrders,
		diffMarkets,
		diffParties,
		diffLimits,
		diffAssets,
		diffDelegations,
		diffEpoch,
		diffVegaTime,
		diffNodes,
		diffNetParams,
		diffProposals,
		diffDeposits,
		diffWithdrawals,
		diffLPs,
		diffStake,
		diffTransfers,
	}

	for _, v := range diffFuncs {
		r := v(dr.coreResult, dr.datanodeResult)
		if r.MatchResult != FullMatch {
			dr.Success = false
		}
		println("completed diff for", r.Key, "with status", r.MatchResult)
		dr.DiffResult = append(dr.DiffResult, r)
	}
	println("completed diff with success?", dr.Success)
}

// diffAccountBalances compares account balances - assuming all accounts should be in both the core snapshot and the datanode.
func diffAccountBalances(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Accounts
	datanode := dn.Accounts

	markets := map[string]struct{}{}
	for _, m := range coreSnapshot.Markets {
		markets[m.Id] = struct{}{}
	}
	markets[""] = struct{}{}

	filteredDN := []*dnproto.AccountBalance{}
	// datanode would have margin and bond accounts for settled markets so need to exclude them
	for _, ab := range datanode {
		if _, ok := markets[ab.MarketId]; ok || ab.Owner == "" {
			filteredDN = append(filteredDN, ab)
		}
	}
	datanode = filteredDN

	sort.Slice(core, func(i, j int) bool {
		return core[i].Owner+core[i].MarketId+core[i].Asset+core[i].Type.String() < core[j].Owner+core[j].MarketId+core[j].Asset+core[j].Type.String()
	})

	dnData := map[string]*dnproto.AccountBalance{}
	for _, ab := range datanode {
		id := ab.Owner + ab.MarketId + ab.Asset + ab.Type.String()
		dnData[id] = ab
	}

	for _, a := range core {
		id := a.Owner + a.MarketId + a.Asset + a.Type.String()
		if d, ok := dnData[id]; ok {
			if a.String() != d.String() && a.Type != vega.AccountType_ACCOUNT_TYPE_EXTERNAL {
				return getValueMismatchStatus("accounts", core, datanode)
			}
		}
	}

	return getSuccessStatus("accounts", core, datanode)
}

// diffOrders compares live orders from core snapshot and datanode.
// NB: parked orders from datanode are excluded in advance.
func diffOrders(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Orders
	datanode := dn.Orders
	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("orders", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		// core may increment UpdatedAt, but if nothing changes it doesn't send an event
		d.UpdatedAt = a.UpdatedAt
		if a.String() != d.String() {
			return getValueMismatchStatus("orders", core, datanode)
		}
	}

	return getSuccessStatus("orders", core, datanode)
}

// diffMarkets compares active markets from core snapshots with the same from datanode.
func diffMarkets(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Markets
	coreIds := map[string]struct{}{}
	for _, m := range core {
		coreIds[m.Id] = struct{}{}
	}

	datanode := []*vega.Market{}
	for _, m := range dn.Markets {
		if _, ok := coreIds[m.Id]; ok {
			datanode = append(datanode, m)
		}
	}

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })
	if len(core) != len(datanode) {
		return getSizeMismatchStatus("markets", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("markets", core, datanode)
		}
	}

	return getSuccessStatus("markets", core, datanode)
}

// diffParties compares parties from core snapshot (i.e. collateral accounts and staking accounts) with datanode.
func diffParties(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Parties
	datanode := dn.Parties

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("parties", core, datanode)
	}

	for i, a := range core {
		if a.String() != datanode[i].String() {
			return getValueMismatchStatus("parties", core, datanode)
		}
	}

	return getSuccessStatus("parties", core, datanode)
}

// diffLimits compares network limits.
func diffLimits(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Limits
	datanode := dn.Limits

	if core.String() != datanode.String() {
		return getSimpleValueMismatchStatus("limits", core.String(), datanode.String())
	}

	return getSimpleSuccessStatus("limits")
}

// assuming all assets ever existed are returned by both.
func diffAssets(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Assets
	datanode := dn.Assets

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("assets", core, datanode)
	}

	for i, a := range core {
		if a.String() != datanode[i].String() {
			return getValueMismatchStatus("assets", core, datanode)
		}
	}

	return getSuccessStatus("assets", core, datanode)
}

// diffDelegations compares the live delegations from the core with the corresponding delegtions from datanode.
// As datanode would return all delegations, need to filter by epochs the core has. For those epochs the state much perfectly match.
func diffDelegations(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Delegations

	epochs := map[string]struct{}{}
	for _, d := range core {
		epochs[d.EpochSeq] = struct{}{}
	}

	datanode := []*vega.Delegation{}
	for _, d := range dn.Delegations {
		if _, ok := epochs[d.EpochSeq]; ok {
			datanode = append(datanode, d)
		}
	}

	sort.Slice(core, func(i, j int) bool {
		ai := core[i]
		aj := core[j]
		return ai.EpochSeq+"_"+ai.NodeId+"_"+ai.Party < aj.EpochSeq+"_"+aj.NodeId+"_"+aj.Party
	})
	sort.Slice(datanode, func(i, j int) bool {
		ai := datanode[i]
		aj := datanode[j]
		return ai.EpochSeq+"_"+ai.NodeId+"_"+ai.Party < aj.EpochSeq+"_"+aj.NodeId+"_"+aj.Party
	})
	if len(core) != len(datanode) {
		return getSizeMismatchStatus("delegations", core, datanode)
	}
	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			getValueMismatchStatus("delegations", core, datanode)
		}
	}

	return getSuccessStatus("delegations", core, datanode)
}

// diffEpoch compares the timestamps of epoch from core snapshot and datanode.
func diffEpoch(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Epoch
	datanode := dn.Epoch

	if core.Seq != datanode.Seq || core.Timestamps.StartTime != datanode.Timestamps.StartTime || core.Timestamps.ExpiryTime != datanode.Timestamps.ExpiryTime {
		return getSimpleValueMismatchStatus("epoch", core.String(), datanode.String())
	}

	return getSimpleSuccessStatus("epoch")
}

// diffVegaTime compares the current vega time on core snapshot and datanode.
func diffVegaTime(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.VegaTime
	datanode := dn.VegaTime
	if core != datanode {
		return getSimpleValueMismatchStatus("vegaTime", strconv.FormatInt(core, 10), strconv.FormatInt(datanode, 10))
	}

	return getSimpleSuccessStatus("vegaTime")
}

// diffNodes compares the validator list on core snapshot and datanode.
// TODO: need to see what happens on a network where a node has been announced to be added in a future epoch - such node would be returned by the
// core snapshot but not by datanode APi.
func diffNodes(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Nodes
	datanode := dn.Nodes

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("nodes", core, datanode)
	}
	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("nodes", core, datanode)
		}
	}
	return getSuccessStatus("nodes", core, datanode)
}

// diffNetParams compares enacted and pending governance proposals from core snapshot and datanode.
func diffNetParams(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.NetParams
	datanode := dn.NetParams
	sort.Slice(core, func(i, j int) bool { return core[i].Key < core[j].Key })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Key < datanode[j].Key })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("netparams", core, datanode)
	}

	for i, a := range core {
		if a.String() != datanode[i].String() {
			return getValueMismatchStatus("netparams", core, datanode)
		}
	}

	return getSuccessStatus("netparams", core, datanode)
}

// diffProposals compares enacted and pending governance proposals from core snapshot and datanode.
func diffProposals(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Proposals
	datanode := dn.Proposals
	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("proposals", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("proposals", core, datanode)
		}
		// if a.Id != d.Id ||
		// 	a.Reference != d.Reference ||
		// 	a.PartyId != d.PartyId ||
		// 	a.State != d.State ||
		// 	a.Timestamp != d.Timestamp ||
		// 	a.Terms.ClosingTimestamp != d.Terms.ClosingTimestamp ||
		// 	a.Terms.EnactmentTimestamp != d.Terms.EnactmentTimestamp {
		// 	return errResult
		// }
		// switch a.Terms.Change.(type) {
		// case *vega.ProposalTerms_NewMarket:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_NewMarket).NewMarket.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_NewMarket).NewMarket.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// case *vega.ProposalTerms_UpdateMarket:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_UpdateMarket).UpdateMarket.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_UpdateMarket).UpdateMarket.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// case *vega.ProposalTerms_NewAsset:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_NewAsset).NewAsset.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_NewAsset).NewAsset.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// case *vega.ProposalTerms_UpdateAsset:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_UpdateAsset).UpdateAsset.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_UpdateAsset).UpdateAsset.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// case *vega.ProposalTerms_NewFreeform:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_NewFreeform).NewFreeform.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_NewFreeform).NewFreeform.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// case *vega.ProposalTerms_UpdateNetworkParameter:
		// 	cCore := a.Terms.Change.(*vega.ProposalTerms_UpdateNetworkParameter).UpdateNetworkParameter.String()
		// 	cDN := d.Terms.Change.(*vega.ProposalTerms_UpdateNetworkParameter).UpdateNetworkParameter.String()
		// 	if cCore != cDN {
		// 		return errResult
		// 	}
		// }
	}
	return getSuccessStatus("proposals", core, datanode)
}

// diffDeposits compares *live* deposits from the core snapshot with the same from datanode.
func diffDeposits(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Deposits
	datanode := []*vega.Deposit{}
	coreDIDs := map[string]struct{}{}
	for _, w := range core {
		coreDIDs[w.Id] = struct{}{}
	}

	// filter only live deposits from datanode
	for _, w := range dn.Deposits {
		if _, ok := coreDIDs[w.Id]; ok {
			datanode = append(datanode, w)
		}
	}

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })
	if len(core) != len(datanode) {
		return getSizeMismatchStatus("deposits", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("deposits", core, datanode)
		}
	}

	return getSuccessStatus("deposits", core, datanode)
}

func keyIntersection[K comparable, V any](mapA map[K]V, mapB map[K]V) []K {
	result := []K{}
	for k := range mapA {
		if _, ok := mapB[k]; ok {
			result = append(result, k)
		}
	}
	return result
}

// diffWithdrawals compares *live* withdarawls from the core snapshot with the same from datanode.
func diffWithdrawals(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Withdrawals
	datanode := dn.Withdrawals

	coreByID := map[string]*vega.Withdrawal{}
	datanodeByID := map[string]*vega.Withdrawal{}

	for _, w := range coreSnapshot.Withdrawals {
		coreByID[w.Id] = w
	}

	for _, w := range dn.Withdrawals {
		datanodeByID[w.Id] = w
	}

	// Only compare if withdrawal is both core and datanode; core may have more as it currently
	// never deletes any, but a datanode without history will initially have none.
	// Issue about core behavior: https://github.com/vegaprotocol/vega/issues/7440
	intersection := keyIntersection(coreByID, datanodeByID)

	for _, id := range intersection {
		if coreByID[id].String() != datanodeByID[id].String() {
			return getValueMismatchStatus("withdrawals", core, datanode)
		}
	}

	return getSuccessStatus("withdrawals", core, datanode)
}

// diffLPs compares liquidity provisions from live markets from core snapshot with the same market LPs in datanode.
// only active and undeployed LPs are compared.
func diffLPs(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Lps
	datanode := []*vega.LiquidityProvision{}
	markets := map[string]struct{}{}
	for _, m := range coreSnapshot.Markets {
		markets[m.Id] = struct{}{}
	}
	for _, ab := range dn.Lps {
		if _, ok := markets[ab.MarketId]; ok {
			datanode = append(datanode, ab)
		}
	}
	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("liquidityProvisions", core, datanode)
	}
	for i, a := range core {
		if a.String() != datanode[i].String() {
			return getValueMismatchStatus("liquidityProvisions", core, datanode)
		}
	}

	return getSuccessStatus("liquidityProvisions", core, datanode)
}

// diffStake compares stake linking from the core and data node. They are expected to be of the same size and perfectly matching at this point.
func diffStake(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Stake
	datanode := dn.Stake
	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("stake", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("stake", core, datanode)
		}
	}
	return getSuccessStatus("stake", core, datanode)
}

// diffTransfers compares live transfers from the core snapshot with datanode.
// All live transfers from the core snapshot must exist on the data node snapshot and match perfectly.
func diffTransfers(coreSnapshot *Result, dn *Result) Status {
	core := coreSnapshot.Transfers
	datanode := []*v1.Transfer{}

	coreIDs := map[string]struct{}{}
	for _, t := range core {
		coreIDs[t.Id] = struct{}{}
	}
	for _, d := range dn.Transfers {
		if _, ok := coreIDs[d.Id]; ok {
			datanode = append(datanode, d)
		}
	}

	sort.Slice(core, func(i, j int) bool { return core[i].Id < core[j].Id })
	sort.Slice(datanode, func(i, j int) bool { return datanode[i].Id < datanode[j].Id })

	if len(core) != len(datanode) {
		return getSizeMismatchStatus("transfers", core, datanode)
	}

	for i, a := range core {
		d := datanode[i]
		if a.String() != d.String() {
			return getValueMismatchStatus("transfers", core, datanode)
		}
	}

	return getSuccessStatus("transfers", core, datanode)
}

func getSuccessStatus[A interface{ String() string }](key string, core, datanode []A) Status {
	return Status{
		Key:         key,
		MatchResult: FullMatch,
		CoreResLen:  len(core),
		DataNodeLen: len(datanode),
	}
}

func getSimpleSuccessStatus(key string) Status {
	return Status{
		Key:         key,
		MatchResult: FullMatch,
		CoreResLen:  1,
		DataNodeLen: 1,
	}
}

func getSizeMismatchStatus[A interface{ String() string }](key string, core, datanode []A) Status {
	return Status{
		Key:         key,
		MatchResult: SizeMismatch,
		CoreRes:     sliceToString(core),
		DatanodeRes: sliceToString(datanode),
		CoreResLen:  len(core),
		DataNodeLen: len(datanode),
	}
}

func getValueMismatchStatus[A interface{ String() string }](key string, core, datanode []A) Status {
	return Status{
		Key:         key,
		MatchResult: ValuesMismatch,
		CoreRes:     sliceToString(core),
		DatanodeRes: sliceToString(datanode),
		CoreResLen:  len(core),
		DataNodeLen: len(datanode),
	}
}

func getSimpleValueMismatchStatus(key string, core, datanode string) Status {
	return Status{
		Key:         key,
		MatchResult: ValuesMismatch,
		CoreRes:     core,
		DatanodeRes: datanode,
		CoreResLen:  1,
		DataNodeLen: 1,
	}
}

func sliceToString[A interface{ String() string }](s []A) string {
	str := ""
	for _, o := range s {
		a := o.String()
		str += a + "\n"
	}
	return str
}
