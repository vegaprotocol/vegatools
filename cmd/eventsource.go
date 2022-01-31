package cmd

import (
	"code.vegaprotocol.io/vegatools/eventsource"
	"github.com/spf13/cobra"
	"time"
)

var (
	eventSourceOpts struct {
		file                  string
		port                  uint
		intervalBetweenBlocks uint
		closeConnection       bool
		logFormat             string
	}

	eventSourceCmd = &cobra.Command{
		Use:   "eventsource",
		Short: "acts as a source of vega events read from a file",
		RunE:  runEventSource,
	}
)

func init() {
	rootCmd.AddCommand(eventSourceCmd)
	eventSourceCmd.Flags().StringVarP(&eventSourceOpts.file, "file", "f", "vega.evt", "name of the file to read events from")
	eventSourceCmd.Flags().UintVarP(&eventSourceOpts.port, "port", "p", 8022, "the port of which to listen for ")
	eventSourceCmd.Flags().UintVarP(&eventSourceOpts.intervalBetweenBlocks, "intervalBetweenBlocks", "i", 1000, "the time interval in milli secs between events being published for each block")
	eventSourceCmd.Flags().BoolVarP(&eventSourceOpts.closeConnection, "closeConnection", "c", false, "close the connection after all events are sent")
	eventSourceCmd.Flags().StringVar(&eventSourceOpts.logFormat, "log-format", "raw", "console data logged in specified format. Allowed values: raw (default), text, json")
}

func runEventSource(cmd *cobra.Command, args []string) error {
	return eventsource.Run(eventSourceOpts.file, eventSourceOpts.port, eventSourceOpts.closeConnection,
		time.Duration(eventSourceOpts.intervalBetweenBlocks)*time.Millisecond, eventSourceOpts.logFormat)
}
