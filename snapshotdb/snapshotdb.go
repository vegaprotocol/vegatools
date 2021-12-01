package snapshotdb

import (
	"fmt"

	"github.com/cosmos/iavl"
	"github.com/syndtr/goleveldb/leveldb/opt"
	db "github.com/tendermint/tm-db"
)

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
	_, err = tree.Load()
	versions := tree.AvailableVersions()

	if versionsOnly {
		// Only output a json version of the number of snapshot versions that are in this database
		fmt.Printf("{ Versions: %d }\n", len(versions))
	} else {
		// Run through all the snapshots and display some details
		fmt.Printf("{ Snapshots: {\n")
		for _, version := range versions {
			v, err := tree.LazyLoadVersion(int64(version))
			if err == nil {
				fmt.Printf("  { Version: %d, Height: %d, Size: %d },\n", v, tree.Height(), tree.Size())
			}
		}
		fmt.Printf(" }\n}\n")
	}
	return nil
}
