package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sahib/brig/bench"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func allBenchmarks() []string {
	names := []string{}

	for _, benchName := range bench.BenchmarkNames() {
		for _, inputName := range bench.InputNames() {
			names = append(names, fmt.Sprintf("%s:%s", benchName, inputName))
		}
	}

	return names
}

func printStats(s bench.Stats) {
	fmt.Println()
	fmt.Println("Time:         ", s.Time.Format(time.RFC3339))
	fmt.Println("CPU Name:     ", s.CPUBrandName)
	fmt.Println("Logical Cores:", s.LogicalCores)
	fmt.Println("Has AESNI:    ", yesify(s.HasAESNI))
	fmt.Println()
}

type benchmarkRun struct {
	Stats   bench.Stats    `json:"stats"`
	Results []bench.Result `json:"results"`
}

func handleIOBench(ctx *cli.Context) error {
	run := benchmarkRun{
		Stats: bench.FetchStats(),
	}

	benchmarks := ctx.StringSlice("bench")
	if len(benchmarks) == 0 {
		log.Infof("running all benchmarks...")
		benchmarks = allBenchmarks()
	}

	isJSON := ctx.Bool("json")
	if !isJSON {
		printStats(run.Stats)
	}

	inputSize, err := humanize.ParseBytes(ctx.String("size"))
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
			Size:        inputSize,
			Encryption:  ctx.String("encryption"),
			Compression: ctx.String("compression"),
		})
	}

	var baselineTiming time.Duration
	var lastSection string

	err = bench.Benchmark(cfgs, func(result bench.Result) {
		section := fmt.Sprintf(
			"%s:%s",
			result.Config.InputName,
			result.Config.BenchName,
		)

		if section != lastSection {
			if !isJSON {
				drawHeading(section)
			}

			// First in list is always the none-none benchmark.
			baselineTiming = result.Took
			lastSection = section
		}

		if !isJSON {
			drawBar(
				result.Name,
				result.Took,
				baselineTiming,
				inputSize,
				result.Throughput,
			)
		}

		run.Results = append(run.Results, result)
	})

	if err != nil {
		return err
	}

	if isJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		enc.Encode(run)
	}

	return nil
}

func drawHeading(heading string) {
	fmt.Println()
	fmt.Println(heading)
	fmt.Println(strings.Repeat("=", len(heading)))
	fmt.Println()
}

func drawBar(name string, took, ref time.Duration, inputSize uint64, throughput float64) {
	w, _, err := terminal.GetSize(1)
	if err != nil {
		w = 100
	}

	perc := float64(ref) / float64(took)

	// take a guess on how big the bar may be.
	var cells int = 50
	if barw := w - 80; barw < cells {
		cells = barw
	}

	fmt.Printf("%-45s [", name)
	for idx := 0; idx < cells; idx++ {
		if idx <= int(perc*float64(cells)) {
			fmt.Printf("=")
		} else {
			fmt.Printf("_")
		}
	}

	fmt.Printf(
		"] %-7.2fMB/s %15s %7.2f%%\n",
		throughput,
		fmt.Sprintf(
			"%.2fMB/%v",
			float64(inputSize)/1000/1000,
			took.Round(time.Millisecond),
		),
		perc*100,
	)
}

func handleIOBenchList(ctx *cli.Context) error {
	for _, name := range allBenchmarks() {
		fmt.Println(name)
	}

	return nil
}
