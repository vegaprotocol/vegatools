package cmd

import (
	"code.vegaprotocol.io/vegatools/perftest"
	"github.com/spf13/cobra"
)

var (
	perfTestOpts struct {
		dataNodeAddr      string
		walletURL         string
		faucetURL         string
		commandsPerSecond int
		runtimeSeconds    int
	}

	perfTestCmd = &cobra.Command{
		Use:   "perftest",
		Short: "perftest runs a constant message load on the network",
		RunE:  runPerfTest,
	}
)

func init() {
	rootCmd.AddCommand(perfTestCmd)
	perfTestCmd.Flags().StringVarP(&perfTestOpts.dataNodeAddr, "address", "a", "", "address of the data node server")
	perfTestCmd.Flags().StringVarP(&perfTestOpts.walletURL, "wallet", "w", "", "address of the wallet server")
	perfTestCmd.Flags().StringVarP(&perfTestOpts.faucetURL, "faucet", "f", "", "address of the faucet server")
	perfTestCmd.MarkFlagRequired("address")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(perfTestOpts.dataNodeAddr,
		perfTestOpts.walletURL,
		perfTestOpts.faucetURL,
		perfTestOpts.commandsPerSecond,
		perfTestOpts.runtimeSeconds)
}
