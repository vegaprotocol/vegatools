# Withdraw tokens from ERC20 Bridge contract

This command is used to impersonate validators in the eyes of the ERC20 Bridge contract, so it allows withdrawing funds. This command ends up being handy when a network has crashed or is being reset but people didn't have time to withdraw their funds.

This requires private keys for all validators, and is meant for testnet `ONLY`.

## Help

```bash
make build
./build/vegatools unsafe_withdraw_impersonate_validators --help
```

## Running

Create the file `validator-privkeys.json`:

```json
[
  "1111111111111111111111111111111111111111111111111111111111111111",
  "2222222222222222222222222222222222222222222222222222222222222222",
  "3333333333333333333333333333333333333333333333333333333333333333"
]
```

Then run:

```bash
make build
./build/vegatools unsafe_withdraw_impersonate_validators \
    --amount 100000 \
    --asset-address 0xaa... \
    --bridge-address 0xbb... \
    --priv-keys validator-privkeys.json
```

Then open a browser and:

1. Open https://www.myetherwallet.com/
1. Click "Access My Wallet"
1. Click "MEW CX", accept terms and conditions, click "Access My Wallet"
1. In the left nav, click "Contract", then "Interact with Contract"
1. Contract address: the address of the ERC20 Bridge contract
1. ABI JSON: Paste content from [MultisigControl: `ERC20_Bridge_Logic_ABI.json`](https://github.com/vegaprotocol/MultisigControl/blob/develop/ropsten_deploy_details/test/ERC20_Bridge_Logic_ABI.json)
1. Select contract function: `withdraw_asset`
1. Copy the command output values into the function parameters list
1. Click "Write"
