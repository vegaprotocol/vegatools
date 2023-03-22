package cmd

import (
	"time"

	"code.vegaprotocol.io/vegatools/benchmark"

	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "A benchmarking tool for vega APIs",
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.PersistentFlags().StringSliceP("urls", "u", []string{"localhost:3007"}, "gRPC host:port of the server to benchmark")
	benchmarkCmd.PersistentFlags().DurationP("timeout", "t", time.Minute, "Timeout for each benchmark test")
	benchmarkCmd.PersistentFlags().IntP("iterations", "i", 1, "Number of iterations to run")
	benchmarkCmd.PersistentFlags().IntP("workers", "w", 1, "Number of concurrent workers to use")
	benchmarkCmd.PersistentFlags().IntP("requests", "q", 100, "Number of requests to send per iteration")

	benchmarkCmd.AddCommand(benchmark.ListOrdersCmd)
}
