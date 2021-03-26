package cmd

import (
	"code.vegaprotocol.io/vegatools/marketDepthViewer"
	"github.com/spf13/cobra"
)

var (
	marketDepthViewerOpts struct {
		market     string
		serverAddr string
	}

	marketDepthViewerCmd = &cobra.Command{
		Use:   "marketdepthviewer",
		Short: "Display the market depth for a single market",
		RunE:  runMarketDepthViewer,
	}
)

func init() {
	rootCmd.AddCommand(marketDepthViewerCmd)
	marketDepthViewerCmd.Flags().StringVarP(&marketDepthViewerOpts.market, "market", "m", "", "name of the market to listen for updates")
	marketDepthViewerCmd.Flags().StringVarP(&marketDepthViewerOpts.serverAddr, "address", "a", "", "address of the grpc server")
	marketDepthViewerCmd.MarkFlagRequired("address")
}

func runMarketDepthViewer(cmd *cobra.Command, args []string) error {
	return marketDepthViewer.Run(marketDepthViewerOpts.serverAddr, marketDepthViewerOpts.market)
}
