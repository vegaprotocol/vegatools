[![Actions Status](https://github.com/vegaprotocol/vegatools/workflows/build/badge.svg)](https://github.com/vegaprotocol/vegatools/actions)

VEGATOOLS
=========

This repo contains a suite of (sometimes) useful tools to use with the vega nodes API.

## How to install
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

### Vega unsafe_withdraw_impersonate_validators

This command is used to impersonate validators in the eyes of bridge into withdrawing funds. This command end up being handy, when a network will crash / is being reset but people didn't have time to withdraw their funds.

This require validators private keys, and is meant for testnet `ONLY`.

Here's an example on how to run it:
```
vegatools unsafe_withdraw_impersonate_validators --amount="1000" --asset-address="0x7e50...8e344" --bridge-address="0x4761...4883" --receiver-address="0xE20c...74EE" --priv-keys="keys.json"
```

This will dump on your terminal the information that you will need to call the withdraw method from the bridge smart contract, you then just need to load the bridge smart contract into MEW, and copy past the details that the command did output to you.
