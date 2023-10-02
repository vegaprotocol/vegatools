package cmd

import (
	"fmt"
	"strings"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vegatools/events"
	"github.com/spf13/cobra"
)

var (
	evt struct {
		evtIn   string
		jsonIn  string
		outPath string
		number  uint64
		types   []int32
		create  bool
	}

	eventsCmd = &cobra.Command{
		Use:   "events",
		Short: "Dump, edit, extract data-node events input file",
		RunE:  parseEvents,
	}
)

func init() {
	help := make([]string, 0, len(eventspb.BusEventType_name)+1)
	help = append(help, "List of events to filter by (default all) - event numbers:")
	defaults := make([]int32, 0, len(eventspb.BusEventType_name))
	for v, name := range eventspb.BusEventType_name {
		if v == 0 || v == 1 {
			continue // 0 is unspecified, 1 == all
		}
		defaults = append(defaults, v)
		help = append(help, fmt.Sprintf("  %d = %s", v, name))
	}
	details := strings.Join(help, "\n")
	rootCmd.AddCommand(eventsCmd)
	eventsCmd.Flags().StringVarP(&evt.evtIn, "event", "e", "", "event file input")
	eventsCmd.Flags().StringVarP(&evt.jsonIn, "json", "j", "", "JSON input file")
	eventsCmd.Flags().StringVarP(&evt.outPath, "out", "o", "", "Output file")
	eventsCmd.Flags().Uint64VarP(&evt.number, "number", "n", 0, "Number of events to parse [default: 0 == all]")
	eventsCmd.Flags().Int32SliceVarP(&evt.types, "types", "t", defaults, details)
}

func parseEvents(cmd *cobra.Command, args []string) error {
	// if our input file is JSON, and we are writing to an output file, flag this run as a create command
	if len(evt.outPath) != 0 && len(evt.jsonIn) != 0 {
		evt.create = true
	}
	return events.Run(
		evt.evtIn,
		evt.jsonIn,
		evt.outPath,
		evt.number,
		evt.types,
		evt.create,
	)
}
