package cmd

import (
	"code.vegaprotocol.io/vegatools/eventpersister"
	"github.com/spf13/cobra"
)

var (
	persistEventOpts struct {
		file       string
		batchSize  uint
		party      string
		market     string
		serverAddr string
		logFormat  string
		reconnect  bool
		types      []string
	}

	persistEventsCmd = &cobra.Command{
		Use:   "persistevents",
		Short: "Persist events from a vega node to a file",
		RunE:  runPersistEvents,
	}
)

func init() {
	rootCmd.AddCommand(persistEventsCmd)
	persistEventsCmd.Flags().StringVarP(&persistEventOpts.file, "file", "f", "vega.evt", "name of the file to persist events in")
	persistEventsCmd.Flags().UintVarP(&persistEventOpts.batchSize, "batch-size", "b", 0, "size of the event stream batch of events")
	persistEventsCmd.Flags().StringVarP(&persistEventOpts.party, "party", "p", "", "name of the party to listen for updates")
	persistEventsCmd.Flags().StringVarP(&persistEventOpts.market, "market", "m", "", "name of the market to listen for updates")
	persistEventsCmd.Flags().StringVarP(&persistEventOpts.serverAddr, "address", "a", "", "address of the grpc server")
	persistEventsCmd.Flags().StringVar(&persistEventOpts.logFormat, "log-format", "raw", "output stream data in specified format. Allowed values: raw (default), text, json")
	persistEventsCmd.Flags().BoolVarP(&persistEventOpts.reconnect, "reconnect", "r", false, "if connection dies, attempt to reconnect")
	persistEventsCmd.Flags().StringSliceVarP(&persistEventOpts.types, "type", "t", nil, "one or more event types to subscribe to (default=ALL)")
	persistEventsCmd.MarkFlagRequired("address")
}

func runPersistEvents(cmd *cobra.Command, args []string) error {
	return eventpersister.Run(
		persistEventOpts.file,
		persistEventOpts.batchSize,
		persistEventOpts.party,
		persistEventOpts.market,
		persistEventOpts.serverAddr,
		persistEventOpts.logFormat,
		persistEventOpts.reconnect,
		persistEventOpts.types,
	)
}
