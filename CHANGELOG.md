# Changelog

## Unreleased

### 🚨 Breaking changes
- [299](https://github.com/vegaprotocol/vegatools/issues/299) - Removed stream, snapshot and checkpoint tools as they are migrated to the core repo

### 🗑️ Deprecation
- [](https://github.com/vegaprotocol/vegatools/pull/) -

### 🛠 Improvements
- [57](https://github.com/vegaprotocol/vegatools/pull/57) - Add changelog and new project boards actions
- [36](https://github.com/vegaprotocol/vegatools/pull/36) - Update code to the last version of the protos repo
- [38](https://github.com/vegaprotocol/vegatools/pull/38) - Added new delegation tool
- [40](https://github.com/vegaprotocol/vegatools/pull/40) - Updated ints to strings
- [41](https://github.com/vegaprotocol/vegatools/pull/41) - Add delegation and epoch to checkpoint
- [42](https://github.com/vegaprotocol/vegatools/pull/42) - Add block height to checkpoint files
- [44](https://github.com/vegaprotocol/vegatools/pull/44) - Update to latest protos changes
- [48](https://github.com/vegaprotocol/vegatools/pull/48) - Update dependencies => snapshot -> checkpoint
- [50](https://github.com/vegaprotocol/vegatools/pull/50) - Update checkpoint types, add new field to dummy
- [45](https://github.com/vegaprotocol/vegatools/pull/45) - Use TradingService in vegatools stream
- [47](https://github.com/vegaprotocol/vegatools/pull/47) - Update proto version
- [52](https://github.com/vegaprotocol/vegatools/pull/52) - Port code to use last version of protos (layout change)
- [67](https://github.com/vegaprotocol/vegatools/pull/67) - Adding tool to display snapshot database information
- [61](https://github.com/vegaprotocol/vegatools/pull/61) - Update module `code.vegaprotocol.io/protos` to `v0.46.0`
- [60](https://github.com/vegaprotocol/vegatools/pull/60) - Update `actions/checkout` action to `v2`
- [59](https://github.com/vegaprotocol/vegatools/pull/59) - Update `hattan/verify-linked-issue-action` commit hash to `70d4b06`
- [58](https://github.com/vegaprotocol/vegatools/pull/58) - Update `Zomzog/changelog-checker commit` hash to `9f2307a`
- [55](https://github.com/vegaprotocol/vegatools/pull/55) - Update `golang.org/x/crypto` commit hash to `ae814b3`
- [54](https://github.com/vegaprotocol/vegatools/pull/54) - Update `module github.com/ethereum/go-ethereum` to `v1.10.13`
- [64](https://github.com/vegaprotocol/vegatools/pull/64) - Add hard coded list of validators to make the display more useful
- [66](https://github.com/vegaprotocol/vegatools/pull/66) - Add check for valid event names
- [69](https://github.com/vegaprotocol/vegatools/pull/69) - Add rewards to checkout in vegatools
- [72](https://github.com/vegaprotocol/vegatools/pull/72) - Update CHANGELOG.md since GH Action implemented and tidy repo
- [73](https://github.com/vegaprotocol/vegatools/pull/73) - Add key rotations checkpoint
- [83](https://github.com/vegaprotocol/vegatools/pull/83) - Update instruction on how to get the real latest version
- [90](https://github.com/vegaprotocol/vegatools/pull/90) - Build and publish docker image
- [96](https://github.com/vegaprotocol/vegatools/pull/96) - Add a tool to save all withdrawals
- [110](https://github.com/vegaprotocol/vegatools/pull/110) - Liquidity commitment viewer
- [113](https://github.com/vegaprotocol/vegatools/pull/113) - Update checkpoint utility to match current state of vega
- [122](https://github.com/vegaprotocol/vegatools/pull/122) - Update protos for new position state message
- [121](https://github.com/vegaprotocol/vegatools/pull/121) - Add option to `snapshotdb` to print a snapshot to a file as JSON
- [140](https://github.com/vegaprotocol/vegatools/issues/140) - Add status column to liquidity commitment viewer
- [156](https://github.com/vegaprotocol/vegatools/issues/156) - Upgraded go build versions to 1.16 + 1.17
- [165](https://github.com/vegaprotocol/vegatools/issues/165) - Adding perftest subcommand
- [184](https://github.com/vegaprotocol/vegatools/issues/184) - Refactoring perftest
- [185](https://github.com/vegaprotocol/vegatools/issues/185) - Adding event rate measuring tool
- [190](https://github.com/vegaprotocol/vegatools/issues/190) - Removed validation time from market proposal
- [197](https://github.com/vegaprotocol/vegatools/issues/197) - Update protos location to their new place inside Vega
- [206](https://github.com/vegaprotocol/vegatools/issues/206) - Move liquidity provision from inside market proposal to it's own proposal
- [208](https://github.com/vegaprotocol/vegatools/issues/208) - User selectable number of markets for load testing 
- [210](https://github.com/vegaprotocol/vegatools/issues/210) - Added delta based market depth display 
- [214](https://github.com/vegaprotocol/vegatools/issues/214) - Increase gRPC receive buffer size 
- [216](https://github.com/vegaprotocol/vegatools/issues/216) - Display block time in eventrate tool to monitor event stream lag 
- [218](https://github.com/vegaprotocol/vegatools/issues/218) - perftest allows multiple LPs and moving of mid price for random orders 
- [222](https://github.com/vegaprotocol/vegatools/issues/222) - Updated market proposal text with renamed fields
- [225](https://github.com/vegaprotocol/vegatools/issues/225) - Add option to dump total event counts
- [227](https://github.com/vegaprotocol/vegatools/issues/227) - Allow configuration of the LP shape
- [230](https://github.com/vegaprotocol/vegatools/issues/230) - Update datanode api use to v2
- [232](https://github.com/vegaprotocol/vegatools/issues/232) - Add batched orders to perftest
- [234](https://github.com/vegaprotocol/vegatools/issues/234) - Add pegged order support to perftest
- [236](https://github.com/vegaprotocol/vegatools/issues/236) - Diff tool introduced to compare core snapshot with data node `API` 
- [237](https://github.com/vegaprotocol/vegatools/issues/237) - Rename of Oracles to Data Sources
- [240](https://github.com/vegaprotocol/vegatools/issues/240) - Add support for filling price levels before perf testing begins
- [247](https://github.com/vegaprotocol/vegatools/issues/247) - Made ganache value optional for perftool
- [251](https://github.com/vegaprotocol/vegatools/issues/251) - Changed code to use wallet V2 and updated protobuf mod to correct version
- [254](https://github.com/vegaprotocol/vegatools/issues/254) - Updated auth to use VWT header value
- [256](https://github.com/vegaprotocol/vegatools/issues/256) - Fix batch orders 
- [258](https://github.com/vegaprotocol/vegatools/issues/258) - Relax difftool comparison of withdrawals
- [260](https://github.com/vegaprotocol/vegatools/issues/260) - Better handling of staking assets in perftool 
- [264](https://github.com/vegaprotocol/vegatools/issues/264) - Eventrate tool can now output a simple report and then exit for use in scripts 
- [266](https://github.com/vegaprotocol/vegatools/issues/266) - Eventrate tool uses correct gov prop call and new option to only initialise the markets 
- [268](https://github.com/vegaprotocol/vegatools/issues/268) - Streamlatency tool to measure latency difference between two different event streams 
- [270](https://github.com/vegaprotocol/vegatools/issues/270) - signingrate tool to benchmark a server to see at what rate they can sign transactions
- [281](https://github.com/vegaprotocol/vegatools/issues/281) - add a toggle to perftest to enable/disable lp users from creating initial orders 
- [288](https://github.com/vegaprotocol/vegatools/issues/288) - stop processing per testing if the number of price levels is greater than the mid price


### 🐛 Fixes
- [78](https://github.com/vegaprotocol/vegatools/pull/78) - Fix build with missing dependency
- [91](https://github.com/vegaprotocol/vegatools/pull/91) - Output of `snapshotdb` is now valid json
- [173](https://github.com/vegaprotocol/vegatools/issues/173) - Add `MarketTracker` to checkpoint parser
- [262](https://github.com/vegaprotocol/vegatools/issues/262) - Update vega dependency, ignore slippage factors in market and last block in epoch
- [279](https://github.com/vegaprotocol/vegatools/issues/279) - Update to work with new datanode API for querying live orders / liquidity provisions
- [286](https://github.com/vegaprotocol/vegatools/pull/286) - Add support for pagination in the accounts and delegations queries to data-node
- [286](https://github.com/vegaprotocol/vegatools/pull/298) - Add support for pagination in the parties query to data-node

## 0.41.1
*2021-08-31*

### 🛠 Improvements
- [25](https://github.com/vegaprotocol/vegatools/pull/25) - Update to latest protos changes
- [27](https://github.com/vegaprotocol/vegatools/pull/27) - Add types flag to make life easier for QA
- [26](https://github.com/vegaprotocol/vegatools/pull/26) - Add checkpoint tool
- [30](https://github.com/vegaprotocol/vegatools/pull/30) - Update module `github.com/ethereum/go-ethereum` to `v1.10.8`
- [31](https://github.com/vegaprotocol/vegatools/pull/31) - Update module `github.com/gdamore/tcell/v2` to `v2.4.0`
- [35](https://github.com/vegaprotocol/vegatools/pull/35) - Update module `google.golang.org/grpc` to `v1.40.0`


## 0.38.0
*2021-06-11*

### 🛠 Improvements
- [22](https://github.com/vegaprotocol/vegatools/pull/22) - Release  version `v0.38.0`


## 0.37.0
*2021-05-28*

### 🛠 Improvements
- [16](https://github.com/vegaprotocol/vegatools/pull/16) - Liquidity monitoring tool
- [18](https://github.com/vegaprotocol/vegatools/pull/18) - Updated instruction for Windows users
- [20](https://github.com/vegaprotocol/vegatools/pull/20) - Updated input index
- [21](https://github.com/vegaprotocol/vegatools/pull/21) - Release version `v0.37.0`

## 0.36.0
*2021-05-15*

### 🛠 Improvements
- [3](https://github.com/vegaprotocol/vegatools/pull/3) - Change dependency name from `api-clients` to api and upgrade to latest version `v0.33.0`
- [4](https://github.com/vegaprotocol/vegatools/pull/4) - Add GitHub Action workflows
- [5](https://github.com/vegaprotocol/vegatools/pull/5) - Adding market viewer tool with scripts to build and run it
- [6](https://github.com/vegaprotocol/vegatools/pull/6) - Add log format option
- [7](https://github.com/vegaprotocol/vegatools/pull/7) - Liquidity monitoring tool
- [9](https://github.com/vegaprotocol/vegatools/pull/9) - Add market stake viewer
- [10](https://github.com/vegaprotocol/vegatools/pull/10) - Improve withdraw
- [11](https://github.com/vegaprotocol/vegatools/pull/11) - Update Vega api version
- [14](https://github.com/vegaprotocol/vegatools/pull/14) - Release version `v0.36.0`
