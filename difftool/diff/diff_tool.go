package diff

import (
	"errors"
	"fmt"
)

// Run takes a snapshot (proto serialised) file path and data node connection string and runs the diff tool.
// Returns nil if no error is found or the error report otherwise.
func Run(snapshotFilePath, datanodeConnection string) error {
	// get data node data
	datanode := newDataNodeClient(datanodeConnection)
	dataNodeResult, err := datanode.Collect()
	if err != nil {
		return err
	}

	// get core snapshot data
	coreSnapshot, err := newSnapshotData(snapshotFilePath)
	if err != nil {
		return err
	}
	coreResult := coreSnapshot.Collect()

	// generate a diff report
	diffReport := newDiffReport(coreResult, dataNodeResult)

	if !diffReport.Success {
		report := fmt.Sprintf("mismatch between core and datanode: %s", diffReport.String())
		return errors.New(report)
	}
	return nil
}
