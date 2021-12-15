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

### Stream

Stream is a simple utility used to connect to a vega validator and listen to ALL events it produce.

Here's an example of how to run it:
```console
vegatools stream --address=n09.testnet.vega.xyz:3002
```

This will listen to all event from this testnet node, run the following commands for a detailed help and filtering
```console
vegatools stream -h
```

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

### Checkpoint

Checkpoint forces the vega node to generate a checkpoint file.

### DelegationViewer

DelegationViewer displays the amount of stake delegated to each of the nodes in the network.

### MarketStakeViewer

MarketStakeViewer displays the amount of stake committed to each market in terms of liquidity provision.

