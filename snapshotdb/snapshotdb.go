package snapshotdb

import (
	"fmt"

	"github.com/cosmos/iavl"
	db "github.com/tendermint/tm-db"
)

// Run is the main entry point for this tool
func Run(dbpath string, versionsOnly bool) error {
	// Attempt to open the database
	db, err := db.NewGoLevelDB("snapshot", dbpath)
	if err != nil {
		fmt.Errorf("Failed to open database located at %s\n", dbpath)
		return err
	}

	tree, err := iavl.NewMutableTree(db, 0)
	_, err = tree.Load()
	versions := tree.AvailableVersions()

	if versionsOnly {
		// Only output a json version of the number of snapshot versions that are in this database
		fmt.Printf("{ Versions: %d }\n", len(versions))
	} else {
		// Run through all the snapshots and display some details
		fmt.Println("    Block Number          Height            Size")
		for _, version := range versions {
			v, err := tree.LazyLoadVersion(int64(version))
			if err == nil {
				fmt.Printf("%16d%16d%16d\n", v, tree.Height(), tree.Size())
			}
		}
	}
	return nil
}
