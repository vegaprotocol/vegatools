package cmd

import (
	"code.vegaprotocol.io/vegatools/marketdepthviewer"
	marketDepthViewer "code.vegaprotocol.io/vegatools/marketdepthviewer"

	"github.com/spf13/cobra"
)

var (
	marketDepthViewerOpts marketdepthviewer.Opts

	marketDepthViewerCmd = &cobra.Command{
		Use:   "marketdepthviewer",
		Short: "Display the market depth for a single market",
		RunE:  runMarketDepthViewer,
	}
)

func init() {
	rootCmd.AddCommand(marketDepthViewerCmd)
	marketDepthViewerCmd.Flags().StringVarP(&marketDepthViewerOpts.Market, "market", "m", "", "name of the market to listen for updates")
	marketDepthViewerCmd.Flags().StringVarP(&marketDepthViewerOpts.ServerAddr, "address", "a", "", "address of the grpc server")
	marketDepthViewerCmd.Flags().BoolVarP(&marketDepthViewerOpts.UseDeltas, "deltas", "d", true, "use deltas instead of snapshots")
	marketDepthViewerCmd.MarkFlagRequired("address")
}

func runMarketDepthViewer(cmd *cobra.Command, args []string) error {
	return marketDepthViewer.Run(marketDepthViewerOpts)
}
