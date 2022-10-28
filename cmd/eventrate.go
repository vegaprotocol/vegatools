package cmd

import (
	"code.vegaprotocol.io/vegatools/eventrate"

	"github.com/spf13/cobra"
)

var (
	eventRateOpts eventrate.Opts
	eventRateCmd  = &cobra.Command{
		Use:   "eventrate",
		Short: "Display the rate in which event bus messages are arriving",
		RunE:  runEventRate,
	}
)

func init() {
	rootCmd.AddCommand(eventRateCmd)
	eventRateCmd.Flags().StringVarP(&eventRateOpts.ServerAddr, "address", "a", "", "address of the grpc server")
	eventRateCmd.Flags().IntVarP(&eventRateOpts.Buckets, "buckets", "b", 10, "number of historic buckets")
	eventRateCmd.Flags().IntVarP(&eventRateOpts.SecondsPerBucket, "secondsperbucket", "s", 1, "number of seconds to record each bucket")
	eventRateCmd.Flags().IntVarP(&eventRateOpts.EventCountDump, "eventcountdump", "e", 0, "dump total event count every x seconds")
	eventRateCmd.MarkFlagRequired("address")
}

func runEventRate(cmd *cobra.Command, args []string) error {
	return eventrate.Run(eventRateOpts)
}
