package cmd

import (
	"code.vegaprotocol.io/vegatools/streamlatency"
	"github.com/spf13/cobra"
)

var (
	streamLatencyOpts streamlatency.Opts
	streamLatencyCmd  = &cobra.Command{
		Use:   "streamlatency",
		Short: "Display the latency difference between two event streams",
		RunE:  runStreamLatency,
	}
)

func init() {
	rootCmd.AddCommand(streamLatencyCmd)
	streamLatencyCmd.Flags().StringVarP(&streamLatencyOpts.ServerAddr1, "address1", "a", "", "address of the first grpc server")
	streamLatencyCmd.Flags().StringVarP(&streamLatencyOpts.ServerAddr2, "address2", "b", "", "address of the second grpc server")
	streamLatencyCmd.Flags().BoolVarP(&streamLatencyOpts.ReportMode, "reportmode", "r", false, "generate report style output")
	streamLatencyCmd.MarkFlagRequired("address1")
	streamLatencyCmd.MarkFlagRequired("address2")
}

func runStreamLatency(cmd *cobra.Command, args []string) error {
	return streamlatency.Run(streamLatencyOpts)
}
