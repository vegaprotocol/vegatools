package powrate

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	vgcrypto "code.vegaprotocol.io/vega/libs/crypto"
)

// Opts are the command line options passed to the sub command
type Opts struct {
	MinPoWLevel       int
	MaxPoWLevel       int
	TestSeconds       int
	SecondsPerBlock   int
	TxPerBlock        int
	DefaultDifficulty int
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

	if opts.DefaultDifficulty < opts.MinPoWLevel || opts.DefaultDifficulty > opts.MaxPoWLevel {
		return fmt.Errorf("default difficulty value must be between the min and max testable values")
	}

	fmt.Printf("Vega PoW Benchmark (Difficulty %d-%d, %d seconds per level)\n", opts.MinPoWLevel, opts.MaxPoWLevel, opts.TestSeconds)
	fmt.Println("|  PoW Difficulty  | Total PoW Count | PoW Operations Per Second |")
	fmt.Println("------------------------------------------------------------------")

	var waitGroup sync.WaitGroup
	var powTimes map[int]float32 = make(map[int]float32)
	for difficulty := opts.MinPoWLevel; difficulty <= opts.MaxPoWLevel; difficulty++ {
		fmt.Printf("|  %-3d             |", difficulty)
		work := make(chan int, 100)
		stop := make(chan int)
		timeout := make(chan int)

		operationsCount := 0
		waitGroup.Add(1)
		go pow(uint(difficulty), work, stop, &waitGroup)
		go wait(opts.TestSeconds, timeout)
		shouldQuit := false
		for !shouldQuit {
			select {
			case <-work:
				operationsCount++
			case <-timeout:
				stop <- 0
				shouldQuit = true
			}
		}
		waitGroup.Wait()
		fmt.Printf(" %-16d| %24d  |\n", operationsCount, operationsCount/opts.TestSeconds)

		// Store the time per calc for use later
		powTimes[difficulty] = float32(opts.TestSeconds) / float32(operationsCount)
	}
	fmt.Println("------------------------------------------------------------------")

	// Loop through all the timings starting at the default difficulty and add up the time
	totalTime := float32(0.0)
	powCalcs := 0
	foundValue := false
	for difficulty := opts.DefaultDifficulty; difficulty <= opts.MaxPoWLevel; difficulty++ {
		for i := 0; i < opts.TxPerBlock; i++ {
			totalTime += powTimes[difficulty]
			if totalTime > float32(opts.SecondsPerBlock) {
				foundValue = true
				break
			}
			powCalcs++
		}
		if foundValue {
			break
		}
	}
	if foundValue {
		fmt.Printf("For an average blocktime of %d seconds\n", opts.SecondsPerBlock)
		fmt.Printf("and a starting difficulty of %d\n", opts.DefaultDifficulty)
		fmt.Printf("and the maximum number of transactions per block of %d\n", opts.TxPerBlock)
		fmt.Printf("the maximum number of PoW calcs per linked block is %d\n", powCalcs)
		fmt.Println("------------------------------------------------------------------")
	} else {
		fmt.Println("not enough data to generate maximum PoW value")
	}

	return nil
}
