package cmd

import (
	"code.vegaprotocol.io/vegatools/eventsource"
	"github.com/spf13/cobra"
	"time"
)

var (
	dataNodeEventSourceOpts struct {
		file                  string
		port                  uint
		intervalBetweenBlocks uint
		closeConnection       bool
		logFormat             string
	}

	dataNodeEventSourceCmd = &cobra.Command{
		Use:   "datanode_eventsource",
		Short: "reads events from a file and send them to a datanode",
		RunE:  runDatanodeEventSource,
	}
)

func init() {
	rootCmd.AddCommand(dataNodeEventSourceCmd)
	dataNodeEventSourceCmd.Flags().StringVarP(&dataNodeEventSourceOpts.file, "file", "f", "vega.evt", "name of the file to read events from")
	dataNodeEventSourceCmd.Flags().UintVarP(&dataNodeEventSourceOpts.port, "port", "p", 3005, "the datanode's listening port ")
	dataNodeEventSourceCmd.Flags().UintVarP(&dataNodeEventSourceOpts.intervalBetweenBlocks, "intervalBetweenBlocks", "i", 1000, "the time interval in milli secs between events being published for each block")
	dataNodeEventSourceCmd.Flags().BoolVarP(&dataNodeEventSourceOpts.closeConnection, "closeConnection", "c", false, "close the connection after all events are sent")
	dataNodeEventSourceCmd.Flags().StringVar(&dataNodeEventSourceOpts.logFormat, "log-format", "raw", "console data logged in specified format. Allowed values: raw (default), text, json")
}

func runDatanodeEventSource(cmd *cobra.Command, args []string) error {
	return eventsource.RunDatanodeEventSource(dataNodeEventSourceOpts.file, dataNodeEventSourceOpts.port, dataNodeEventSourceOpts.closeConnection,
		time.Duration(dataNodeEventSourceOpts.intervalBetweenBlocks)*time.Millisecond, dataNodeEventSourceOpts.logFormat)
}
