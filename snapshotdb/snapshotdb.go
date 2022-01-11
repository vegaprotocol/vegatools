package snapshotdb

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/iavl"
	"github.com/syndtr/goleveldb/leveldb/opt"
	db "github.com/tendermint/tm-db"
)

// SnapshotData is a representation of the information we an scrape from the avl tree
type SnapshotData struct {
	Version int64 `json:"version"`
	Height  int64 `json:"height"`
	Size    int64 `json:"size"`
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
		j.Snapshots = append(j.Snapshots, SnapshotData{
			Version: v,
			Height:  int64(tree.Height()),
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

// Run is the main entry point for this tool
func Run(dbpath string, versionsOnly bool) error {
	// Attempt to open the database
	options := &opt.Options{
		ErrorIfMissing: true,
		ReadOnly:       true,
	}
	db, err := db.NewGoLevelDBWithOpts("snapshot", dbpath, options)
	if err != nil {
		return fmt.Errorf("failed to open database located at %s : %w", dbpath, err)
	}

	tree, err := iavl.NewMutableTree(db, 0)
	if err != nil {
		return err
	}

	if _, err := tree.Load(); err != nil {
		return err
	}
	versions := tree.AvailableVersions()

	if versionsOnly {
		return displayNumberOfVersions(len(versions))
	}
	return displaySnapshotData(tree, versions)
}
