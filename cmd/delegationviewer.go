package cmd

import (
	delegationViewer "code.vegaprotocol.io/vegatools/delegationviewer"
	"github.com/spf13/cobra"
)

var (
	delegationViewerOpts struct {
		serverAddr  string
		updateDelay uint
	}

	delegationViewerCmd = &cobra.Command{
		Use:   "delegationviewer",
		Short: "Display the delegation values for a vega network",
		RunE:  runDelegationViewer,
	}
)

func init() {
	rootCmd.AddCommand(delegationViewerCmd)
	delegationViewerCmd.Flags().StringVarP(&delegationViewerOpts.serverAddr, "address", "a", "", "address of the grpc server")
	delegationViewerCmd.Flags().UintVarP(&delegationViewerOpts.updateDelay, "delay", "d", 5, "delay in seconds between updates")
	delegationViewerCmd.MarkFlagRequired("address")
}

func runDelegationViewer(cmd *cobra.Command, args []string) error {
	return delegationViewer.Run(delegationViewerOpts.serverAddr, delegationViewerOpts.updateDelay)
}
