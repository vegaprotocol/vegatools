package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"code.vegaprotocol.io/vegatools/withdraw"
	"github.com/spf13/cobra"
)

var (
	unsafeWithdrawImpersonateValiatorsOpts struct {
		privKeysFile            string
		ethereumReceiverAddress string
		ethereumAssetAddress    string
		ethereumBridgeAddress   string
		amount                  string
	}

	unsafeWithdrawImpersonateValiatorsCmd = &cobra.Command{
		Use:   "unsafe_withdraw_impersonate_validators",
		Short: "Uses the validators private keys to sign a payload to withdraw funds",
		RunE:  runUnsafeWithdrawImpersonateValidators,
	}
)

func init() {
	rootCmd.AddCommand(unsafeWithdrawImpersonateValiatorsCmd)
	unsafeWithdrawImpersonateValiatorsCmd.Flags().
		StringVarP(&unsafeWithdrawImpersonateValiatorsOpts.privKeysFile, "priv-keys", "p", "", "The path to a file contains the private keys of the validators")
	unsafeWithdrawImpersonateValiatorsCmd.Flags().
		StringVarP(&unsafeWithdrawImpersonateValiatorsOpts.ethereumReceiverAddress, "receiver-address", "r", "", "The ethereum address to which we want the bridge to send the funds to")
	unsafeWithdrawImpersonateValiatorsCmd.Flags().
		StringVarP(&unsafeWithdrawImpersonateValiatorsOpts.ethereumAssetAddress, "asset-address", "a", "", "The ethereum address of the erc20 token to be withdrawn")
	unsafeWithdrawImpersonateValiatorsCmd.Flags().
		StringVarP(&unsafeWithdrawImpersonateValiatorsOpts.ethereumBridgeAddress, "bridge-address", "b", "", "The ethereum address of the erc20 vega bridge")
	unsafeWithdrawImpersonateValiatorsCmd.Flags().
		StringVarP(&unsafeWithdrawImpersonateValiatorsOpts.amount, "amount", "m", "", "The amount of funds to be withdrawn")

	unsafeWithdrawImpersonateValiatorsCmd.MarkFlagRequired("priv-keys")
	unsafeWithdrawImpersonateValiatorsCmd.MarkFlagRequired("receiver-address")
	unsafeWithdrawImpersonateValiatorsCmd.MarkFlagRequired("asset-address")
	unsafeWithdrawImpersonateValiatorsCmd.MarkFlagRequired("bridge-address")
	unsafeWithdrawImpersonateValiatorsCmd.MarkFlagRequired("amount")
}

func runUnsafeWithdrawImpersonateValidators(cmd *cobra.Command, args []string) error {
	// try to read priv keys
	data, err := ioutil.ReadFile(
		unsafeWithdrawImpersonateValiatorsOpts.privKeysFile)
	if err != nil {
		return fmt.Errorf("unable to read priv-key file, %w", err)
	}

	privKeys := []string{}
	err = json.Unmarshal(data, &privKeys)
	if err != nil {
		return fmt.Errorf("unable to unmarshal priv-keys, %w", err)
	}

	return withdraw.UnsafeWithdrawImpersonateValidators(
		privKeys,
		unsafeWithdrawImpersonateValiatorsOpts.ethereumReceiverAddress,
		unsafeWithdrawImpersonateValiatorsOpts.ethereumAssetAddress,
		unsafeWithdrawImpersonateValiatorsOpts.ethereumBridgeAddress,
		unsafeWithdrawImpersonateValiatorsOpts.amount,
	)
}
