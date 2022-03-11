package checkpoint

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	checkpoint "code.vegaprotocol.io/protos/vega/checkpoint/v1"
	events "code.vegaprotocol.io/protos/vega/events/v1"

	"github.com/golang/protobuf/proto"
)

var (
	// ErrCheckpointFileEmpty obviously means the checkpoint file was empty
	ErrCheckpointFileEmpty = errors.New("given checkpoint file is empty or unreadable")
	// ErrMissingOutFile no output file name argument provided
	ErrMissingOutFile = errors.New("output file not specified")
)

// Run ... the main entry point of the command
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
	cp := &checkpoint.Checkpoint{}
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
	cp, h, err := d.CheckpointData() // get the data as checkpoint
	if err != nil {
		fmt.Printf("Could not convert dummy to checkpoint data to write to file: %+v\n", err)
		return err
	}
	if err := writeCheckpoint(cp, h, cpF); err != nil {
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
	a := &all{}
	if err := a.FromJSON(data); err != nil {
		fmt.Printf("Could not unmarshal input: %+v\n", err)
		return err
	}
	out, h, err := a.CheckpointData()
	if err != nil {
		fmt.Printf("Could not generate checkpoint data: %+v\n", err)
		return err
	}
	hash := hex.EncodeToString(Hash(h))
	n, err := of.Write(out)
	if err != nil {
		fmt.Printf("Failed to write output to file: %+v\n", err)
		return err
	}
	fmt.Printf("Successfully wrote %d bytes to file %s\n", n, outF)
	fmt.Printf("Hash for checkpoint is %s\n", hash)
	return nil
}

func writeCheckpoint(data, h []byte, outF string) error {
	hash := hex.EncodeToString(Hash(h))
	of, err := os.Create(outF)
	if err != nil {
		fmt.Printf("Failed to create output file %s: %+v\n", outF, err)
		return err
	}
	defer func() {
		_ = of.Close()
	}()
	n, err := of.Write(data)
	if err != nil {
		fmt.Printf("Failed to write output to file '%s': %+v\n", outF, err)
		return err
	}
	fmt.Printf("Successfully wrote %d bytes to file %s\n", n, outF)
	fmt.Printf("Checkpoint hash is %s\n", hash)
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
		}
		fmt.Printf("Could not write to stderr: %+v\n", err)
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

func unmarshalAll(cp *checkpoint.Checkpoint) (*all, error) {
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
	if ret.Delegate, err = unmarshalDelegate(cp); err != nil {
		return nil, err
	}
	if ret.Epoch, err = unmarshalEpoch(cp); err != nil {
		return nil, err
	}
	if ret.Block, err = unmarshalBlock(cp); err != nil {
		return nil, err
	}
	if ret.Rewards, err = unmarshalRewards(cp); err != nil {
		return nil, err
	}
	if ret.Banking, err = unmarshalBanking(cp); err != nil {
		return nil, err
	}
	if ret.Validators, err = unmarshalValidators(cp); err != nil {
		return nil, err
	}
	if ret.Staking, err = unmarshalStaking(cp); err != nil {
		return nil, err
	}
	if ret.MultisigControl, err = unmarshalMultisigControl(cp); err != nil {
		return nil, err
	}
	return ret, nil
}

func unmarshalStaking(cp *checkpoint.Checkpoint) (*checkpoint.Staking, error) {
	p := &checkpoint.Staking{}
	if err := proto.Unmarshal(cp.Staking, p); err != nil {
		return nil, err
	}
	return p, nil
}

func unmarshalMultisigControl(cp *checkpoint.Checkpoint) (*checkpoint.MultisigControl, error) {
	p := &checkpoint.MultisigControl{}
	if err := proto.Unmarshal(cp.MultisigControl, p); err != nil {
		return nil, err
	}
	return p, nil
}

func unmarshalGovernance(cp *checkpoint.Checkpoint) (*checkpoint.Proposals, error) {
	p := &checkpoint.Proposals{}
	if err := proto.Unmarshal(cp.Governance, p); err != nil {
		return nil, err
	}
	return p, nil
}

func unmarshalAssets(cp *checkpoint.Checkpoint) (*checkpoint.Assets, error) {
	a := &checkpoint.Assets{}
	if err := proto.Unmarshal(cp.Assets, a); err != nil {
		return nil, err
	}
	return a, nil
}

func unmarshalCollateral(cp *checkpoint.Checkpoint) (*checkpoint.Collateral, error) {
	c := &checkpoint.Collateral{}
	if err := proto.Unmarshal(cp.Collateral, c); err != nil {
		return nil, err
	}
	return c, nil
}

func unmarshalNetParams(cp *checkpoint.Checkpoint) (*checkpoint.NetParams, error) {
	n := &checkpoint.NetParams{}
	if err := proto.Unmarshal(cp.NetworkParameters, n); err != nil {
		return nil, err
	}
	return n, nil
}

func unmarshalDelegate(cp *checkpoint.Checkpoint) (*checkpoint.Delegate, error) {
	d := &checkpoint.Delegate{}
	if err := proto.Unmarshal(cp.Delegation, d); err != nil {
		return nil, err
	}
	return d, nil
}

func unmarshalEpoch(cp *checkpoint.Checkpoint) (*events.EpochEvent, error) {
	e := &events.EpochEvent{}
	if err := proto.Unmarshal(cp.Epoch, e); err != nil {
		return nil, err
	}
	return e, nil
}

func unmarshalBlock(cp *checkpoint.Checkpoint) (*checkpoint.Block, error) {
	b := &checkpoint.Block{}
	if err := proto.Unmarshal(cp.Block, b); err != nil {
		return nil, err
	}
	return b, nil
}

func unmarshalRewards(cp *checkpoint.Checkpoint) (*checkpoint.Rewards, error) {
	r := &checkpoint.Rewards{}
	if err := proto.Unmarshal(cp.Rewards, r); err != nil {
		return nil, err
	}
	return r, nil
}

func unmarshalBanking(cp *checkpoint.Checkpoint) (*checkpoint.Banking, error) {
	kr := &checkpoint.Banking{}
	if err := proto.Unmarshal(cp.Banking, kr); err != nil {
		return nil, err
	}
	return kr, nil
}

func unmarshalValidators(cp *checkpoint.Checkpoint) (*checkpoint.Validators, error) {
	kr := &checkpoint.Validators{}
	if err := proto.Unmarshal(cp.Validators, kr); err != nil {
		return nil, err
	}
	return kr, nil
}
