package cmd

import (
	"code.vegaprotocol.io/vegatools/checkpoint"

	"github.com/spf13/cobra"
)

var (
	cp struct {
		inPath   string
		outPath  string
		format   string
		validate bool
		create   bool
		dummy    bool
	}

	checkpointCmd = &cobra.Command{
		Use:   "checkpoint",
		Short: "Make checkpoint human-readable, or generate checkpoint from human readable format",
		RunE:  parseCheckpoint,
	}
)

func init() {
	rootCmd.AddCommand(checkpointCmd)
	checkpointCmd.Flags().StringVarP(&cp.inPath, "file", "f", "", "input file to parse")
	checkpointCmd.Flags().StringVarP(&cp.outPath, "out", "o", "", "output file to write to [default is STDOUT]")
	checkpointCmd.Flags().StringVarP(&cp.format, "format", "F", "", "output format [default is JSON]")
	checkpointCmd.Flags().BoolVarP(&cp.validate, "validate", "v", false, "validate contents of the checkpoint file")
	checkpointCmd.Flags().BoolVarP(&cp.create, "generate", "g", false, "input is human readable, generate checkpoint file")
	checkpointCmd.Flags().BoolVarP(&cp.dummy, "dummy", "d", false, "generate a dummy file [added for debugging, but could be useful]")
	checkpointCmd.MarkFlagRequired("file")
}

func parseCheckpoint(cmd *cobra.Command, args []string) error {
	return checkpoint.Run(
		cp.inPath,
		cp.outPath,
		cp.format,
		cp.create,
		cp.validate,
		cp.dummy,
	)
}
