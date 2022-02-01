package cmd

import (
	"code.vegaprotocol.io/vegatools/eventsource"
	"github.com/spf13/cobra"
	"time"
)

var (
	grpcEventSourceOpts struct {
		file                  string
		port                  uint
		intervalBetweenBlocks uint
		closeConnection       bool
		logFormat             string
	}

	grpcEventSourceCmd = &cobra.Command{
		Use:   "grpc_eventsource",
		Short: "serves all the events in the given file via the Core grpc events API",
		RunE:  runEventSource,
	}
)

func init() {
	rootCmd.AddCommand(grpcEventSourceCmd)
	grpcEventSourceCmd.Flags().StringVarP(&grpcEventSourceOpts.file, "file", "f", "vega.evt", "name of the file to read events from")
	grpcEventSourceCmd.Flags().UintVarP(&grpcEventSourceOpts.port, "port", "p", 3002, "the port of which to listen for ")
	grpcEventSourceCmd.Flags().UintVarP(&grpcEventSourceOpts.intervalBetweenBlocks, "intervalBetweenBlocks", "i", 1000, "the time interval in milli secs between events being published for each block")
	grpcEventSourceCmd.Flags().BoolVarP(&grpcEventSourceOpts.closeConnection, "closeConnection", "c", false, "close the connection after all events are sent")
	grpcEventSourceCmd.Flags().StringVar(&grpcEventSourceOpts.logFormat, "log-format", "raw", "console data logged in specified format. Allowed values: raw (default), text, json")
}

func runEventSource(cmd *cobra.Command, args []string) error {
	return eventsource.RunGrpcEventSource(grpcEventSourceOpts.file, grpcEventSourceOpts.port, grpcEventSourceOpts.closeConnection,
		time.Duration(grpcEventSourceOpts.intervalBetweenBlocks)*time.Millisecond, grpcEventSourceOpts.logFormat)
}
