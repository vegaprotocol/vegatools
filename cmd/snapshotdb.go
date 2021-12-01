package cmd

import (
	"code.vegaprotocol.io/vegatools/snapshotdb"
	"github.com/spf13/cobra"
)

var (
	snapshotDBOpts struct {
		databasePath     string
		versionCountOnly bool
	}

	snapshotDBCmd = &cobra.Command{
		Use:   "snapshotdb",
		Short: "Displays information about the snapshot database",
		RunE:  runSnapshotDBCmd,
	}
)

func init() {
	rootCmd.AddCommand(snapshotDBCmd)
	snapshotDBCmd.Flags().StringVarP(&snapshotDBOpts.databasePath, "db-path", "d", "", "path to the goleveldb database folder")
	snapshotDBCmd.Flags().BoolVarP(&snapshotDBOpts.versionCountOnly, "versions", "v", false, "display the number of stored versions")
	streamCmd.MarkFlagRequired("db-path")
}

func runSnapshotDBCmd(cmd *cobra.Command, args []string) error {
	return snapshotdb.Run(snapshotDBOpts.databasePath, snapshotDBOpts.versionCountOnly)
}
