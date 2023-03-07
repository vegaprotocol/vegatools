package cmd

import (
	"code.vegaprotocol.io/vegatools/powrate"

	"github.com/spf13/cobra"
)

var (
	powRateOpts powrate.Opts
	powRateCmd  = &cobra.Command{
		Use:   "powrate",
		Short: "Display the rate at which the current machine can process proof of work operations",
		RunE:  runPowRate,
	}
)

func init() {
	rootCmd.AddCommand(powRateCmd)
	powRateCmd.Flags().IntVarP(&powRateOpts.MinPoWLevel, "minimumpowdifficulty", "a", 1, "lowest PoW difficulty to test")
	powRateCmd.Flags().IntVarP(&powRateOpts.MaxPoWLevel, "maximumpowdifficulty", "b", 20, "highest PoW difficulty to test")
	powRateCmd.Flags().IntVarP(&powRateOpts.TestSeconds, "testseconds", "s", 10, "length of time in seconds to test each difficulty level")
}

func runPowRate(cmd *cobra.Command, args []string) error {
	return powrate.Run(powRateOpts)
}
