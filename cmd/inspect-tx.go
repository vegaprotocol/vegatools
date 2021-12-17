package cmd

import (
	"fmt"

	commandspb "code.vegaprotocol.io/protos/vega/commands/v1"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

var (
	rawTransaction []byte
	inspectTxCmd   = &cobra.Command{
		Use:   "inspect-tx",
		Short: "Inspect a raw Vega transaction",
		RunE:  runInspectTx,
	}
)

func init() {
	rootCmd.AddCommand(inspectTxCmd)
	inspectTxCmd.Flags().BytesBase64VarP(&rawTransaction, "tx", "t", []byte(""), "Base64 encoding of the raw Vega transaction to decode")
	inspectTxCmd.MarkFlagRequired("tx")
}

func runInspectTx(cmd *cobra.Command, args []string) error {
	var tx = &commandspb.Transaction{}
	marshaler := jsonpb.Marshaler{
		Indent: "   ",
	}

	if err := proto.Unmarshal(rawTransaction, tx); err != nil {
		return fmt.Errorf("Couldn't unmarshal transaction: %w", err)
	}

	g, err := marshaler.MarshalToString(tx)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal transaction: %w", err)
	}

	fmt.Println(g)
	return nil
}
