package cmd

import (
	"code.vegaprotocol.io/vegatools/pullvotes"
	"github.com/spf13/cobra"
)

var (
	pullVotesOpts struct {
		nodeAddress string
		from, to    uint64
	}

	pullVotesCmd = &cobra.Command{
		Use:   "pull_votes",
		Short: "Pull all votes from a tendermint chain",
		RunE:  runPullVotes,
	}
)

func init() {
	rootCmd.AddCommand(pullVotesCmd)
	pullVotesCmd.Flags().StringVarP(&pullVotesOpts.nodeAddress, "address", "a", "", "tendermint node")
	pullVotesCmd.Flags().Uint64VarP(&pullVotesOpts.from, "from", "f", 0, "start block")
	pullVotesCmd.Flags().Uint64VarP(&pullVotesOpts.to, "to", "t", 0, "end block")

	pullVotesCmd.MarkFlagRequired("address")
	pullVotesCmd.MarkFlagRequired("to")
}

func runPullVotes(cmd *cobra.Command, args []string) error {
	return pullvotes.Start(
		pullVotesOpts.from, pullVotesOpts.to, pullVotesOpts.nodeAddress)
}
