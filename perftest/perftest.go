package perftest

import (
	"flag"
	"fmt"
)

// Run is the main function of `perftest` package
func Run(dataNodeAddr string) error {
	flag.Parse()

	if len(dataNodeAddr) <= 0 {
		return fmt.Errorf("error: missing grpc server address")
	}

	return nil
}
