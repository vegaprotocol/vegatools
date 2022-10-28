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

	opts perftest.Opts
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
	perfTestCmd.Flags().IntVarP(&opts.MarketCount, "markets", "m", 1, "number of markets to create and use")
	perfTestCmd.Flags().IntVarP(&opts.Voters, "voters", "v", 3, "number of accounts to assign voting power")
	perfTestCmd.Flags().IntVarP(&opts.LPOrdersPerSide, "lporders", "l", 3, "number of orders per side in the LP shape")
	perfTestCmd.Flags().IntVarP(&opts.BatchSize, "batchsize", "b", 0, "set size of batch orders")
	perfTestCmd.Flags().IntVarP(&opts.PeggedOrders, "peggedorders", "o", 0, "set number of pegged orders to load the market with")
	perfTestCmd.Flags().BoolVarP(&opts.MoveMid, "movemidprice", "p", false, "allow the mid price we place orders around to move randomly")
	perfTestCmd.MarkFlagRequired("address")
	perfTestCmd.MarkFlagRequired("wallet")
	perfTestCmd.MarkFlagRequired("faucet")
	perfTestCmd.MarkFlagRequired("ganache")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(opts)
}
