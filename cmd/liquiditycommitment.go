package cmd

import (
	liquidityCommitment "code.vegaprotocol.io/vegatools/liquiditycommitment"
	"github.com/spf13/cobra"
)

var (
	liquidityCommitmentOpts struct {
		market     string
		serverAddr string
	}

	liquidityCommitmentCmd = &cobra.Command{
		Use:   "liquiditycommitment",
		Short: "Display the liquidity commitment for a given market",
		RunE:  runLiquidityCommitment,
	}
)

func init() {
	rootCmd.AddCommand(liquidityCommitmentCmd)
	liquidityCommitmentCmd.Flags().StringVarP(&liquidityCommitmentOpts.market, "market", "m", "", "name of the market to monitor")
	liquidityCommitmentCmd.Flags().StringVarP(&liquidityCommitmentOpts.serverAddr, "address", "a", "", "address of the grpc server")
	liquidityCommitmentCmd.MarkFlagRequired("address")
}

func runLiquidityCommitment(cmd *cobra.Command, args []string) error {
	return liquidityCommitment.Run(liquidityCommitmentOpts.serverAddr,
		liquidityCommitmentOpts.market)
}
