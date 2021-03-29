package cmd

import (
	"code.vegaprotocol.io/vegatools/stream"

	"github.com/spf13/cobra"
)

var (
	streamOpts struct {
		batchSize  uint
		party      string
		market     string
		serverAddr string
		logFormat  string
	}

	streamCmd = &cobra.Command{
		Use:   "stream",
		Short: "Stream events from vega node",
		RunE:  runStream,
	}
)

func init() {
	rootCmd.AddCommand(streamCmd)
	streamCmd.Flags().UintVarP(&streamOpts.batchSize, "batch-size", "b", 0, "size of the event stream batch of events")
	streamCmd.Flags().StringVarP(&streamOpts.party, "party", "p", "", "name of the party to listen for updates")
	streamCmd.Flags().StringVarP(&streamOpts.market, "market", "m", "", "name of the market to listen for updates")
	streamCmd.Flags().StringVarP(&streamOpts.serverAddr, "address", "a", "", "address of the grpc server")
	streamCmd.Flags().StringVar(&streamOpts.logFormat, "log-format", "raw", "output stream data in specified format. Allowed values: raw (default), text, json")
	streamCmd.MarkFlagRequired("address")
}

func runStream(cmd *cobra.Command, args []string) error {
	return stream.Run(streamOpts.batchSize, streamOpts.party, streamOpts.market, streamOpts.serverAddr, streamOpts.logFormat)
}
