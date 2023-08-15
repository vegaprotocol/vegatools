package diff

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	snapshot "code.vegaprotocol.io/vega/protos/vega/snapshot/v1"
	db "github.com/cometbft/cometbft-db"
	"github.com/cosmos/iavl"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"google.golang.org/protobuf/proto"
)

// SnapshotData is a representation of the information we an scrape from the avl tree
type SnapshotData struct {
	Version int64  `json:"version"`
	Height  uint64 `json:"height"`
	Size    int64  `json:"size"`
}

func displaySnapshotData(tree *iavl.MutableTree, versions []int) error {
	j := struct {
		Snapshots []SnapshotData `json:"snapshots"`
	}{}

	for _, version := range versions {

		v, err := tree.LazyLoadVersion(int64(version))
		if err != nil {
			return err
		}

		_, blockHeight, _ := getAllPayloads(tree)
		j.Snapshots = append(j.Snapshots, SnapshotData{
			Version: v,
			Height:  blockHeight,
			Size:    tree.Size(),
		})
	}

	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func getAllPayloads(tree *iavl.MutableTree) ([]*snapshot.Payload, uint64, error) {
	payloads := []*snapshot.Payload{}
	var err error
	var blockHeight uint64
	tree.Iterate(func(key []byte, val []byte) (stop bool) {
		p := &snapshot.Payload{}
		err = proto.Unmarshal(val, p)
		if err != nil {
			return true
		}

		// grab block-height while we're here
		switch dt := p.Data.(type) {
		case *snapshot.Payload_AppState:
			blockHeight = dt.AppState.Height
		}

		payloads = append(payloads, p)
		return false
	})

	return payloads, blockHeight, err
}

func writeSnapshotAsJSON(tree *iavl.MutableTree, outputPath string) error {
	// traverse the tree and get the payloads

	payloads, _, err := getAllPayloads(tree)
	if err != nil {
		return err
	}

	f, _ := os.Create(outputPath)
	defer f.Close()

	w := bufio.NewWriter(f)
	m := jsonpb.Marshaler{Indent: "    "}

	for _, p := range payloads {
		s, _ := m.MarshalToString(p)
		w.WriteString(s)
	}

	w.Flush()
	fmt.Println("snapshot payloads written to:", outputPath)
	return nil
}

// writeSnapshotAsProtobuf saves the snapshot as a binary slice of payloads which is more useful for loading when comparing against datanode.
func writeSnapshotAsProtobuf(tree *iavl.MutableTree, outputPath string) error {
	// traverse the tree and get the payloads
	payloads, _, err := getAllPayloads(tree)
	if err != nil {
		return err
	}

	f, _ := os.Create(outputPath)
	defer f.Close()

	w := bufio.NewWriter(f)

	chunk := &snapshot.Chunk{Data: payloads}
	bytes, err := proto.Marshal(chunk)
	if err != nil {
		return err
	}

	w.WriteString(string(bytes))

	w.Flush()
	fmt.Println("snapshot payloads written to:", outputPath)
	return nil
}

func displayNumberOfVersions(versions int) error {
	j := struct {
		Versions int64 `json:"n_versions"`
	}{
		Versions: int64(versions),
	}

	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// SnapshotRun is the main entry point for this tool
func SnapshotRun(dbpath string, versionsOnly bool, outputPath string, heightToOutput int64, outputFormat string) error {
	// Attempt to open the database
	options := &opt.Options{
		ErrorIfMissing: true,
		ReadOnly:       true,
	}
	db, err := db.NewGoLevelDBWithOpts("snapshot", dbpath, options)
	if err != nil {
		return fmt.Errorf("failed to open database located at %s : %w", dbpath, err)
	}

	tree, err := iavl.NewMutableTree(db, 0, false)
	if err != nil {
		return err
	}

	if _, err := tree.Load(); err != nil {
		return err
	}
	versions := tree.AvailableVersions()

	switch {
	case len(outputPath) != 0:

		// find the tree version for the heigh

		for i := len(versions) - 1; i > -1; i-- {
			version := versions[i]

			_, err := tree.LazyLoadVersion(int64(version))
			if err != nil {
				return err
			}

			_, blockHeight, _ := getAllPayloads(tree)

			// either a height wasn't specified so we take the latest
			if heightToOutput == 0 || blockHeight == uint64(heightToOutput) {
				fmt.Println("found snapshot for block-height", blockHeight)
				if outputFormat == "json" {
					return writeSnapshotAsJSON(tree, outputPath)
				}
				if outputFormat == "proto" {
					return writeSnapshotAsProtobuf(tree, outputPath)
				}
				return errors.New("unknown output format requested")
			}
		}
		return fmt.Errorf("could not find snapshot for height %d", heightToOutput)

	case versionsOnly:
		return displayNumberOfVersions(len(versions))
	default:
		return displaySnapshotData(tree, versions)
	}
}
