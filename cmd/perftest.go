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
	perfTestCmd.Flags().StringVarP(&opts.TokenKeysFile, "tokenkeys", "t", "", "path to api token keys file")
	perfTestCmd.Flags().IntVarP(&opts.CommandsPerSecond, "cps", "c", 100, "commands per second")
	perfTestCmd.Flags().IntVarP(&opts.RuntimeSeconds, "runtime", "r", 60, "runtime in seconds")
	perfTestCmd.Flags().IntVarP(&opts.NormalUserCount, "normalusers", "n", 10, "number of normal users to send commands with")
	perfTestCmd.Flags().IntVarP(&opts.LpUserCount, "lpusers", "u", 3, "number of lp users to provide liquidity")
	perfTestCmd.Flags().IntVarP(&opts.MarketCount, "markets", "m", 1, "number of markets to create and use")
	perfTestCmd.Flags().IntVarP(&opts.Voters, "voters", "v", 3, "number of accounts to assign voting power")
	perfTestCmd.Flags().IntVarP(&opts.BatchSize, "batchsize", "b", 0, "set size of batch orders")
	perfTestCmd.Flags().IntVarP(&opts.PeggedOrders, "peggedorders", "p", 0, "set number of pegged orders to load the market with")
	perfTestCmd.Flags().IntVarP(&opts.PriceLevels, "pricelevels", "L", 100, "number of price levels per side")
	perfTestCmd.Flags().IntVarP(&opts.SLAUpdateSeconds, "slaupdate", "S", 10, "number of seconds between SLA updates")
	perfTestCmd.Flags().IntVarP(&opts.SLAPriceLevels, "slapricelevels", "l", 3, "number of price levels for SLA orders")
	perfTestCmd.Flags().IntVarP(&opts.StopOrders, "stoporders", "d", 0, "number of stop orders per user")
	perfTestCmd.Flags().IntVarP(&opts.AMMs, "amms", "A", 0, "number of automatic market making orders to place")
	perfTestCmd.Flags().Int64VarP(&opts.StartingMidPrice, "startingmidprice", "s", 10000, "mid price to use at the start")
	perfTestCmd.Flags().BoolVarP(&opts.MoveMid, "movemidprice", "M", false, "allow the mid price we place orders around to move randomly")
	perfTestCmd.Flags().BoolVarP(&opts.FillPriceLevels, "fillpricelevel", "F", false, "place an order at every available price level")
	perfTestCmd.Flags().BoolVarP(&opts.InitialiseOnly, "initialiseonly", "i", false, "initialise everything and then exit")
	perfTestCmd.Flags().BoolVarP(&opts.DoNotInitialise, "donotinitialise", "I", false, "skip the initialise steps")
	perfTestCmd.Flags().BoolVarP(&opts.UseLPsForOrders, "uselpsfororders", "U", true, "allow lp users to place orders during setup")
	perfTestCmd.MarkFlagRequired("address")
	perfTestCmd.MarkFlagRequired("wallet")
	perfTestCmd.MarkFlagRequired("faucet")
	perfTestCmd.MarkFlagRequired("tokenkeys")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(opts)
}
