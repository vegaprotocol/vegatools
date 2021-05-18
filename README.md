[![Go](https://github.com/vegaprotocol/vegatools/actions/workflows/go.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/go.yml)
[![Go coverage](https://github.com/vegaprotocol/vegatools/actions/workflows/go-coverage.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/go-coverage.yml)
[![YAML lint](https://github.com/vegaprotocol/vegatools/actions/workflows/yml-lint.yml/badge.svg)](https://github.com/vegaprotocol/vegatools/actions/workflows/yml-lint.yml)

VEGATOOLS
=========

This repo contains a suite of (sometimes) useful tools to use with the vega nodes API.

## How to install

To download and build the project your local machine must have the golang tool-chain and the gcc compiler installed. For Windows users you can get the latest and easy to install version of gcc from here:
https://nuwen.net/mingw.html. After extracting the archive, make sure the bin folder is in your PATH.

You can install this program by running the following go install command:
```console
go install code.vegaprotocol.io/vegatools@latest
```

## Available tools

### Vega stream

Vega stream is a simple utility used to connect to a vega validator and listen to ALL events it produce.

Here's an example of how to run it:
```console
vegatools stream --address="n09.testnet.vega.xyz:3002"
```

This will listen to all event from this testnet node, run the following commands for a detailed help and filtering
```
vegatools stream -h
```
