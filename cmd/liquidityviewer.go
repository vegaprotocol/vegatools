package cmd

import (
	liquidityViewer "code.vegaprotocol.io/vegatools/liquidityviewer"
	"github.com/spf13/cobra"
)

var (
	liquidityViewerOpts struct {
		party      string
		market     string
		serverAddr string
	}

	liquidityViewerCmd = &cobra.Command{
		Use:   "liquidityviewer",
		Short: "Display the liquidity commitment and orders for a given party and market",
		RunE:  runLiquidityViewer,
	}
)

func init() {
	rootCmd.AddCommand(liquidityViewerCmd)
	liquidityViewerCmd.Flags().StringVarP(&liquidityViewerOpts.party, "party", "p", "", "name of the party to monitor")
	liquidityViewerCmd.Flags().StringVarP(&liquidityViewerOpts.market, "market", "m", "", "name of the market to monitor")
	liquidityViewerCmd.Flags().StringVarP(&liquidityViewerOpts.serverAddr, "address", "a", "", "address of the grpc server")
	liquidityViewerCmd.MarkFlagRequired("address")
}

func runLiquidityViewer(cmd *cobra.Command, args []string) error {
	return liquidityViewer.Run(liquidityViewerOpts.serverAddr,
		liquidityViewerOpts.market,
		liquidityViewerOpts.party)
}
