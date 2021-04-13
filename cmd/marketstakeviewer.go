package cmd

import (
	"code.vegaprotocol.io/vegatools/marketstakeviewer"

	"github.com/spf13/cobra"
)

var (
	marketStakeViewerOpts struct {
		serverAddr string
	}

	marketStakeViewerCmd = &cobra.Command{
		Use:   "marketstakeviewer",
		Short: "Display market stake info for all markets",
		RunE:  runMarketStakeViewer,
	}
)

func init() {
	rootCmd.AddCommand(marketStakeViewerCmd)
	marketStakeViewerCmd.Flags().StringVarP(&marketStakeViewerOpts.serverAddr, "address", "a", "", "address of the grpc server (host:port)")
	marketStakeViewerCmd.MarkFlagRequired("address")
}

func runMarketStakeViewer(cmd *cobra.Command, args []string) error {
	return marketstakeviewer.Run(marketStakeViewerOpts.serverAddr)
}
