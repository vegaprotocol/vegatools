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
		ganacheURL        string
		commandsPerSecond int
		runtimeSeconds    int
		userCount         int
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
	perfTestCmd.Flags().StringVarP(&perfTestOpts.ganacheURL, "ganache", "g", "", "address of the ganache server")
	perfTestCmd.Flags().IntVarP(&perfTestOpts.commandsPerSecond, "cps", "c", 100, "commands per second")
	perfTestCmd.Flags().IntVarP(&perfTestOpts.runtimeSeconds, "runtime", "r", 60, "runtime in seconds")
	perfTestCmd.Flags().IntVarP(&perfTestOpts.userCount, "users", "u", 10, "number of users to send commands with")
	perfTestCmd.MarkFlagRequired("address")
	perfTestCmd.MarkFlagRequired("wallet")
	perfTestCmd.MarkFlagRequired("faucet")
	perfTestCmd.MarkFlagRequired("ganache")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(perfTestOpts.dataNodeAddr,
		perfTestOpts.walletURL,
		perfTestOpts.faucetURL,
		perfTestOpts.ganacheURL,
		perfTestOpts.commandsPerSecond,
		perfTestOpts.runtimeSeconds,
		perfTestOpts.userCount)
}
