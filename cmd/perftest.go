package cmd

import (
	"code.vegaprotocol.io/vegatools/perftest"
	"github.com/spf13/cobra"
)

var (
	perfTestCmd = &cobra.Command{
		Use:   "perftest",
		Short: "perftest runs a constant message load on the network",
		RunE:  runPerfTest,
	}

	opts perftest.PerfTestOpts
)

func init() {
	rootCmd.AddCommand(perfTestCmd)
	perfTestCmd.Flags().StringVarP(&opts.DataNodeAddr, "address", "a", "", "address of the data node server")
	perfTestCmd.Flags().StringVarP(&opts.WalletURL, "wallet", "w", "", "address of the wallet server")
	perfTestCmd.Flags().StringVarP(&opts.FaucetURL, "faucet", "f", "", "address of the faucet server")
	perfTestCmd.Flags().StringVarP(&opts.GanacheURL, "ganache", "g", "", "address of the ganache server")
	perfTestCmd.Flags().IntVarP(&opts.CommandsPerSecond, "cps", "c", 100, "commands per second")
	perfTestCmd.Flags().IntVarP(&opts.RuntimeSeconds, "runtime", "r", 60, "runtime in seconds")
	perfTestCmd.Flags().IntVarP(&opts.UserCount, "users", "u", 10, "number of users to send commands with")
	perfTestCmd.MarkFlagRequired("address")
	perfTestCmd.MarkFlagRequired("wallet")
	perfTestCmd.MarkFlagRequired("faucet")
	perfTestCmd.MarkFlagRequired("ganache")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(opts)
}
