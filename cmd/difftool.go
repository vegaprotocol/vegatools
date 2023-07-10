package cmd

import (
	"os"
	"strings"

	"code.vegaprotocol.io/vegatools/difftool/diff"
	"code.vegaprotocol.io/vegatools/snapshotdb"
	"github.com/spf13/cobra"
)

var (
	diffToolOpts struct {
		snapshotDatabasePath string
		heightToOutput       int64
		datanode             string
	}

	diffToolCmd = &cobra.Command{
		Use:   "difftool",
		Short: "Compare the state of a core snapshot with datanode API",
		RunE:  runDiffToolCmd,
	}
)

func init() {
	rootCmd.AddCommand(diffToolCmd)
	diffToolCmd.Flags().StringVarP(&diffToolOpts.snapshotDatabasePath, "snap-db-path", "s", "", "path to the goleveldb database folder")
	diffToolCmd.Flags().Int64VarP(&diffToolOpts.heightToOutput, "block-height", "r", 0, "block-height of the snapshot to dump")
	diffToolCmd.Flags().StringVarP(&diffToolOpts.datanode, "datanode", "d", "", "datanode url")
	diffToolCmd.MarkFlagRequired("snap-db-path")
	diffToolCmd.MarkFlagRequired("datanode")
}

func runDiffToolCmd(cmd *cobra.Command, args []string) error {
	temp := os.TempDir()
	if !strings.HasSuffix(temp, string(os.PathSeparator)) {
		temp = temp + string(os.PathSeparator)
	}
	println(temp)
	snapshotPath := temp + "snapshot.dat"

	err := snapshotdb.Run(diffToolOpts.snapshotDatabasePath, false, snapshotPath, diffToolOpts.heightToOutput, "proto")
	defer os.Remove(snapshotPath)
	if err != nil {
		return err
	}

	return diff.Run(snapshotPath, diffToolOpts.datanode)
}
