package cmd

import (
	"code.vegaprotocol.io/vegatools/snapshotdb"
	"github.com/spf13/cobra"
)

var (
	snapshotDBOpts struct {
		databasePath     string
		outputPath       string
		heightToOutput   int64
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
	snapshotDBCmd.Flags().StringVarP(&snapshotDBOpts.outputPath, "out", "o", "", "file to write JSON to")
	snapshotDBCmd.Flags().Int64VarP(&snapshotDBOpts.heightToOutput, "block-height", "r", 0, "block-height of the snapshot to dump")
	snapshotDBCmd.Flags().BoolVarP(&snapshotDBOpts.versionCountOnly, "versions", "v", false, "display the number of stored versions")
	snapshotDBCmd.MarkFlagRequired("db-path")
}

func runSnapshotDBCmd(cmd *cobra.Command, args []string) error {
	return snapshotdb.Run(snapshotDBOpts.databasePath, snapshotDBOpts.versionCountOnly, snapshotDBOpts.outputPath, snapshotDBOpts.heightToOutput)
}
