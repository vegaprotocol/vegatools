[![Go](https://github.com/vegaprotocol/vegatools/actions/workflows/go.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/go.yml)
[![Go coverage](https://github.com/vegaprotocol/vegatools/actions/workflows/go-coverage.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/go-coverage.yml)
[![YAML lint](https://github.com/vegaprotocol/vegatools/actions/workflows/yml-lint.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/yml-lint.yml)

VEGATOOLS
=========

This repo contains a suite of (sometimes) useful tools to use with the vega nodes API.

## How to install

You can install this program by running the following go install command:
```console
// To get the latest stable release version
go install code.vegaprotocol.io/vegatools@latest
// To get the version most in line with the core develop branch
go install code.vegaprotocol.io/vegatools@develop
```
Make sure that your `CGO_ENABLED` environment variable is set to 0. This can be checked using this command:
```console
go env
```

It can be set correctly by:
```console
go env -w CGO_ENABLED=0
```


## Available tools

### MarketDepthViewer

MarketDepthViewer is a utility that will display the market depth of a given market running on a node.

The basic command to run it is:
```console
vegatools marketdepthviewer --address=n09.testnet.vega.xyz:3002
```

If there are multiple markets on a node it will display a list of them at startup and allow the user to select the one they wish to view. Pressing `q` will close the app.

### LiquidityViewer

LiquidityViewer is a utility that displays the liquidity commitment of a user on a particular market.

The basic command to run it is:
```console
vegatools liquidityviewer --address=n09.testnet.vega.xyz:3002
```

If there are multiple markets on the node it will list them and allow the user to select one. If there are multiple users on that market supplying liquidity then it will also list those and allow the user to select one. Pressing `q` will closer the app.

### MarketStakeViewer

MarketStakeViewer is a utility that displays the current state of liquidity provision for all markets running on a node.

The basic command to run it is:
```console
vegatools marketstakeviewer --address=n09.testnet.vega.xyz:3002
```

### DelegationViewer

DelegationViewer displays the amount of stake delegated to each of the nodes in the network.

### MarketStakeViewer

MarketStakeViewer displays the amount of stake committed to each market in terms of liquidity provision.

### LiquidityCommitment

LiquidityCommitment displays all of the liquidity providers for a given market including the fee and commitment amount.

### PerfTest

This creates a market and a set of users and then generates a consistent flow of transactions to the market over a given length of time to allow for performance testing and statistics recording.

### EventRate
This listens to an unfiltered event bus stream and reports the number of events arriving per time bucket (default 1 second) and the amount of network bandwidth it used to receive them. The bucket length and the number of historic buckets it uses to generate the average values can be set on the commandline. 

### PoWRate
This runs a benchmark using the same proof of work algorithm used during the signing process to prevent spam. It reports back the number of transactions per second the machine can process and can be used to help set the proof of work difficulty value for a network. It will also allow uses to get a feel for the rate in which a wallet service will be able to process and forward on transactions if run on the same machine.