package checkpoint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	snapshot "code.vegaprotocol.io/protos/vega/snapshot/v1"

	"github.com/golang/protobuf/proto"
)

var (
	ErrCheckpointFileEmpty = errors.New("given checkpoint file is empty or unreadable")
	ErrMissingOutFile      = errors.New("output file not specified")
)

func Run(inFile, outFile, format string, generate, validate, dummy bool) error {
	if generate && outFile == "" {
		fmt.Println("No output file specified")
		return ErrMissingOutFile
	}
	// generate some files to play with
	if dummy {
		return generateDummy(inFile, outFile)
	}
	data, err := ioutil.ReadFile(inFile)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return ErrCheckpointFileEmpty
	}
	if generate {
		return generateCheckpoint(data, outFile)
	}
	cp := &snapshot.Checkpoint{}
	if err := proto.Unmarshal(data, cp); err != nil {
		return err
	}
	parsed, err := unmarshalAll(cp)
	if err != nil {
		return err
	}
	// print output at the end
	defer func() {
		printParsed(parsed, err != nil)
	}()
	if validate {
		if err = parsed.CheckAssetsCollateral(); err != nil {
			return err
		}
	}
	if outFile != "" {
		if err = writeOut(parsed, outFile); err != nil {
			return err
		}
	}
	return nil
}

func generateDummy(cpF, JSONFname string) error {
	d := dummy()
	cp, err := d.SnapshotData() // get the data as snapshot
	if err != nil {
		fmt.Printf("Could not convert dummy to snapshot data to write to file: %+v\n", err)
		return err
	}
	if err := generateCheckpoint(cp, cpF); err != nil {
		fmt.Printf("Error writing checkpoint data to file '%s': %+v\n", cpF, err)
		return err
	}
	if JSONFname == "" {
		return nil
	}
	if err := writeOut(d, JSONFname); err != nil {
		fmt.Printf("Error writing JSON file '%s' from dummy: %+v\n", JSONFname, err)
		return err
	}
	return nil
}

func generateCheckpoint(data []byte, outF string) error {
	of, err := os.Create(outF)
	if err != nil {
		fmt.Printf("Failed to create output file %s: %+v\n", outF, err)
		return err
	}
	defer func() {
		_ = of.Close()
	}()
	a := all{}
	if err := a.FromJSON(data); err != nil {
		fmt.Printf("Could not unmarshal input: %+v\n", err)
		return err
	}
	out, err := a.SnapshotData()
	if err != nil {
		fmt.Printf("Could not generate snapshot data: %+v\n", err)
		return err
	}
	n, err := of.Write(out)
	if err != nil {
		fmt.Printf("Failed to write output to file: %+v\n", err)
		return err
	}
	fmt.Printf("Successfully wrote %d bytes to file %s\n", n, outF)
	return nil
}

func printParsed(a *all, isErr bool) {
	data, err := a.JSON()
	if err != nil {
		fmt.Printf("Failed to marshal data to JSON: %+v\n", err)
		return
	}
	if isErr {
		if _, err := os.Stderr.WriteString(string(data)); err == nil {
			return
		} else {
			fmt.Printf("Could not write to stderr: %+v\n", err)
		}
	}
	fmt.Printf("Output:\n%s\n", string(data))
}

func writeOut(a *all, path string) error {
	data, err := a.JSON()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	n, err := f.Write(data)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to %s\n", n, path)
	return nil
}

func unmarshalAll(cp *snapshot.Checkpoint) (*all, error) {
	ret := &all{}
	var err error
	if ret.Governance, err = unmarshalGovernance(cp); err != nil {
		return nil, err
	}
	if ret.Assets, err = unmarshalAssets(cp); err != nil {
		return nil, err
	}
	if ret.Collateral, err = unmarshalCollateral(cp); err != nil {
		return nil, err
	}
	if ret.NetParams, err = unmarshalNetParams(cp); err != nil {
		return nil, err
	}
	return ret, nil
}

func unmarshalGovernance(cp *snapshot.Checkpoint) (*snapshot.Proposals, error) {
	p := &snapshot.Proposals{}
	if err := proto.Unmarshal(cp.Governance, p); err != nil {
		return nil, err
	}
	return p, nil
}

func unmarshalAssets(cp *snapshot.Checkpoint) (*snapshot.Assets, error) {
	a := &snapshot.Assets{}
	if err := proto.Unmarshal(cp.Assets, a); err != nil {
		return nil, err
	}
	return a, nil
}

func unmarshalCollateral(cp *snapshot.Checkpoint) (*snapshot.Collateral, error) {
	c := &snapshot.Collateral{}
	if err := proto.Unmarshal(cp.Collateral, c); err != nil {
		return nil, err
	}
	return c, nil
}

func unmarshalNetParams(cp *snapshot.Checkpoint) (*snapshot.NetParams, error) {
	n := &snapshot.NetParams{}
	if err := proto.Unmarshal(cp.NetworkParameters, n); err != nil {
		return nil, err
	}
	return n, nil
}
