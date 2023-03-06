package cmd

import (
	"code.vegaprotocol.io/vegatools/signingrate"

	"github.com/spf13/cobra"
)

var (
	signingRateOpts signingrate.Opts
	signingRateCmd  = &cobra.Command{
		Use:   "signingrate",
		Short: "Display the rate at which the current machine can sign transactions",
		RunE:  runSigningRate,
	}
)

func init() {
	rootCmd.AddCommand(signingRateCmd)
	signingRateCmd.Flags().IntVarP(&signingRateOpts.MinPoWLevel, "minimumpowdifficulty", "a", 1, "lowest PoW difficulty to test")
	signingRateCmd.Flags().IntVarP(&signingRateOpts.MaxPoWLevel, "maximumpowdifficulty", "b", 20, "highest PoW difficulty to test")
	signingRateCmd.Flags().IntVarP(&signingRateOpts.TestSeconds, "testseconds", "s", 10, "length of time in seconds to test each difficulty level")
}

func runSigningRate(cmd *cobra.Command, args []string) error {
	return signingrate.Run(signingRateOpts)
}
