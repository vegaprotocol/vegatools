package cmd

import (
	"code.vegaprotocol.io/vegatools/withdraw"
	"github.com/spf13/cobra"
)

var (
	pullWithdrawOpts struct {
		out         string
		nodeAddress string
	}

	pullWithdrawCmd = &cobra.Command{
		Use:   "pull_withdraw",
		Short: "Pull all withdraw from all parties for a network",
		RunE:  runPullWithdraw,
	}
)

func init() {
	rootCmd.AddCommand(pullWithdrawCmd)
	pullWithdrawCmd.Flags().StringVarP(&pullWithdrawOpts.out, "out", "o", "withdraws.json", "the path to a file to store all withdrawal bundles")
	pullWithdrawCmd.Flags().StringVarP(&pullWithdrawOpts.nodeAddress, "address", "a", "", "address of the grpc server")

	pullWithdrawCmd.MarkFlagRequired("address")
}

func runPullWithdraw(cmd *cobra.Command, args []string) error {
	return withdraw.PullWithdraw(pullWithdrawOpts.nodeAddress, pullWithdrawOpts.out)
}
