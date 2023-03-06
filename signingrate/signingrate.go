package signingrate

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	vgcrypto "code.vegaprotocol.io/vega/libs/crypto"
)

// Opts are the command line options passed to the sub command
type Opts struct {
	MinPoWLevel int
	MaxPoWLevel int
	TestSeconds int
}

func getRandomBlockHash() string {
	var bytes [64]byte
	for i := 0; i < 64; i++ {
		bytes[i] = byte('A' + rand.Intn(26))
	}
	return string(bytes[:])
}

func pow(difficulty uint, work, stop chan int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	tid := "testtransactionid"
	for {
		select {
		case <-stop:
			close(work)
			return
		default:
			_, _, err := vgcrypto.PoW(getRandomBlockHash(), tid, difficulty, vgcrypto.Sha3)
			if err != nil {
				fmt.Println(err)
				return
			}
			work <- 0
		}
	}
}

func wait(seconds int, timer chan int) {
	time.Sleep(time.Duration(seconds) * time.Second)
	timer <- 0
}

// Run is the main function of `signingrate` package
func Run(opts Opts) error {
	if opts.MinPoWLevel > opts.MaxPoWLevel {
		return fmt.Errorf("min difficulty is higher than the max difficulty")
	}

	if opts.MinPoWLevel < 1 {
		return fmt.Errorf("minimum difficulty must be > 0")
	}

	if opts.MaxPoWLevel > 255 {
		return fmt.Errorf("maximum difficulty must be < 256")
	}

	if opts.TestSeconds < 1 && opts.TestSeconds > 100 {
		return fmt.Errorf("test seconds must be > 1 and < 100")
	}

	fmt.Println("Vega Wallet Transaction Signing Benchmark")
	fmt.Println("|  POW Difficulty  | Total Signing Count |  Signings Per Second  |")
	fmt.Println("------------------------------------------------------------------")

	var waitGroup sync.WaitGroup
	for difficulty := opts.MinPoWLevel; difficulty <= opts.MaxPoWLevel; difficulty++ {
		fmt.Printf("|  %-3d             |", difficulty)
		work := make(chan int, 100)
		stop := make(chan int)
		timeout := make(chan int)

		signCount := 0
		waitGroup.Add(1)
		go pow(uint(difficulty), work, stop, &waitGroup)
		go wait(opts.TestSeconds, timeout)
		shouldQuit := false
		for !shouldQuit {
			select {
			case <-work:
				signCount++
			case <-timeout:
				stop <- 0
				shouldQuit = true
			}
		}
		waitGroup.Wait()
		fmt.Printf(" %-20d| %20d  |\n", signCount, signCount/opts.TestSeconds)
	}
	fmt.Println("------------------------------------------------------------------")
	return nil
}
