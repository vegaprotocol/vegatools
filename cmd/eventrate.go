package cmd

import (
	"code.vegaprotocol.io/vegatools/eventrate"

	"github.com/spf13/cobra"
)

var (
	eventRateOpts struct {
		serverAddr string
	}

	eventRateCmd = &cobra.Command{
		Use:   "eventrate",
		Short: "Display the rate in which event bus messages are arriving",
		RunE:  runEventRate,
	}
)

func init() {
	rootCmd.AddCommand(eventRateCmd)
	eventRateCmd.Flags().StringVarP(&eventRateOpts.serverAddr, "address", "a", "", "address of the grpc server")
	eventRateCmd.MarkFlagRequired("address")
}

func runEventRate(cmd *cobra.Command, args []string) error {
	return eventrate.Run(eventRateOpts.serverAddr)
}
