package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sahib/brig/bench"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	defaultBenchmarks = []string{
		// TODO: define default benchmarks. (just all?)
		"null:ten",
		"null:random",
		"mio-reader:ten",
		"mio-reader:random",
		"mio-writer:ten",
		"mio-writer:random",
		"brig-stage-mem:random",
		"brig-cat-mem:random",
		"brig-stage-ipfs:random",
		"brig-cat-ipfs:random",
	}
)

func handleIOBench(ctx *cli.Context) error {
	benchmarks := ctx.StringSlice("bench")

	if len(benchmarks) == 0 {
		benchmarks = defaultBenchmarks
	}

	size, err := humanize.ParseBytes(ctx.String("size"))
	if err != nil {
		return err
	}

	log.SetLevel(log.WarnLevel)

	cfgs := []bench.Config{}
	for _, benchmark := range benchmarks {
		benchSplit := strings.SplitN(benchmark, ":", 2)

		benchInput := "ten"
		benchName := benchSplit[0]
		if len(benchSplit) >= 2 {
			benchInput = benchSplit[1]
		}

		cfgs = append(cfgs, bench.Config{
			BenchName:   benchName,
			InputName:   benchInput,
			Size:        size,
			Encryption:  ctx.String("encryption"),
			Compression: ctx.String("compression"),
		})
	}

	var baselineTiming time.Duration
	var lastSection string

	return bench.Benchmark(cfgs, func(result bench.Result) {
		section := fmt.Sprintf(
			"%s => %s",
			result.Config.InputName,
			result.Config.BenchName,
		)

		if section != lastSection {
			drawHeading(section)

			// First in list is always the none-none benchmark.
			baselineTiming = result.Took
			lastSection = section
		}

		benchName := fmt.Sprintf("enc-%s:zip-%s", result.Encryption, result.Compression)
		drawBar(benchName, result.Took, baselineTiming, size)
	})
}

func drawHeading(heading string) {
	fmt.Println()
	fmt.Println(heading)
	fmt.Println(strings.Repeat("=", len(heading)))
	fmt.Println()
}

func drawBar(name string, took, ref time.Duration, inputSize uint64) {
	perc := float64(ref) / float64(took)
	const cells = 78

	fmt.Printf("%-60s [", name)
	for idx := 0; idx < cells; idx++ {
		if idx <= int(perc*cells) {
			fmt.Printf("=")
		} else {
			fmt.Printf(" ")
		}
	}

	throughput := float64(inputSize) / (float64(took) / float64(time.Second)) / (1024 * 1024)
	fmt.Printf("] %.2f MB/s (%.2f%%)\n", throughput, perc*100)
}
