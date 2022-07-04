package cmd

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

var defaultEthSyncTimeout = time.Second * 10

var (
	ethereumOpts struct {
		address string
		chainID int64
	}

	ethereumCmd = &cobra.Command{
		Use:   "ethereum",
		Short: "Ethereum allows to call certain smart contracts on Ethereum network.",
	}
)

func init() {
	rootCmd.AddCommand(ethereumCmd)

	ethereumCmd.PersistentFlags().StringVar(&ethereumOpts.address, "address", "", "address of the Ethereum network")
	ethereumCmd.PersistentFlags().Int64Var(&ethereumOpts.chainID, "chain-id", 0, "address of the Ethereum network")
	ethereumCmd.MarkPersistentFlagRequired("address")
	ethereumCmd.MarkPersistentFlagRequired("chain-id")
}

func printEthereumTx(tx *types.Transaction) error {
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal transaction to JSON: %w", err)
	}

	fmt.Printf("Transaction: %s", txJSON)

	return nil
}
