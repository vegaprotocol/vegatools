package cmd

import (
	"code.vegaprotocol.io/vegatools/perftest"
	"github.com/spf13/cobra"
)

var (
	perfTestOpts struct {
		dataNodeAddr string
	}

	perfTestCmd = &cobra.Command{
		Use:   "perftest",
		Short: "Perftest runs a constant message load on the network",
		RunE:  runPerfTest,
	}
)

func init() {
	rootCmd.AddCommand(perfTestCmd)
	streamCmd.Flags().StringVarP(&perfTestOpts.dataNodeAddr, "address", "a", "", "address of the data node server")
	streamCmd.MarkFlagRequired("address")
}

func runPerfTest(cmd *cobra.Command, args []string) error {
	return perftest.Run(perfTestOpts.dataNodeAddr)
}
