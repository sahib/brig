package bench

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/repo/setup"
	log "github.com/sirupsen/logrus"
)

// Config define how the benchmarks are run.
type Config struct {
	InputName   string `json:"input_name"`
	BenchName   string `json:"bench_name"`
	Size        uint64 `json:"size"`
	Encryption  string `json:"encryption"`
	Compression string `json:"compression"`
	Samples     int    `json:"samples"`
}

// Result is the result of a single benchmark run.
type Result struct {
	Name        string        `json:"name"`
	Config      Config        `json:"config"`
	Encryption  string        `json:"encryption"`
	Compression string        `json:"compression"`
	Took        time.Duration `json:"took"`
	Throughput  float64       `json:"throughput"`
}

// buildHints handles wildcards for compression and/or encryption.
// If no wildcards are specified, we just take what is set in `cfg`.
func buildHints(cfg Config) []hints.Hint {
	encIsWildcard := cfg.Encryption == "*"
	zipIsWildcard := cfg.Compression == "*"

	if encIsWildcard && zipIsWildcard {
		return hints.AllPossibleHints()
	}

	if encIsWildcard {
		hs := []hints.Hint{}
		for _, encAlgo := range hints.ValidEncryptionHints() {
			hs = append(hs, hints.Hint{
				CompressionAlgo: hints.CompressionHint(cfg.Compression),
				EncryptionAlgo:  hints.EncryptionHint(encAlgo),
			})
		}

		return hs
	}

	if zipIsWildcard {
		hs := []hints.Hint{}
		for _, zipAlgo := range hints.ValidCompressionHints() {
			hs = append(hs, hints.Hint{
				CompressionAlgo: hints.CompressionHint(zipAlgo),
				EncryptionAlgo:  hints.EncryptionHint(cfg.Encryption),
			})
		}

		return hs
	}

	return []hints.Hint{{
		CompressionAlgo: hints.CompressionHint(cfg.Compression),
		EncryptionAlgo:  hints.EncryptionHint(cfg.Encryption),
	}}
}

func sortHints(hs []hints.Hint) []hints.Hint {
	sort.Slice(hs, func(i, j int) bool {
		return hs[i].Less(hs[j])
	})

	// sorts in-place, but also return for ease of use.
	return hs
}

func benchmarkSingle(cfg Config, fn func(result Result), ipfsPath string) error {
	in, err := InputByName(cfg.InputName, cfg.Size)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := ByName(cfg.BenchName, ipfsPath)
	if err != nil {
		return err
	}

	defer out.Close()

	for _, hint := range sortHints(buildHints(cfg)) {
		supportsHints := out.SupportHints()
		if !supportsHints {
			// Indicate in output that nothing was encrypted or compressed.
			hint.CompressionAlgo = hints.CompressionNone
			hint.EncryptionAlgo = hints.EncryptionNone
		}

		if hint.CompressionAlgo == hints.CompressionGuess {
			// NOTE: We do not benchmark guessing here.
			// Simply reason is that we do not know from the output
			// which algorithm was actually used.
			continue
		}

		var tookSum time.Duration

		for seed := uint64(0); seed < uint64(cfg.Samples); seed++ {
			r, err := in.Reader(seed)
			if err != nil {
				return err
			}

			v, err := in.Verifier()
			if err != nil {
				return err
			}

			took, err := out.Bench(hint, r, v)
			if err != nil {
				return err
			}

			tookSum += took

			// Most write-only benchmarks cannot be verified, since
			// we modify the stream and the verifier checks that the stream
			// is equal to the input. Most read tests involve the same logic
			// as writing though, so the writer has to work for that.
			if out.CanBeVerified() {
				if missing := v.MissingBytes(); missing != 0 {
					log.Warnf("not all or too much data received in verify: %d", missing)
				}
			}
		}

		avgTook := tookSum / time.Duration(cfg.Samples)

		// NOTE: We take the configured size, we don't check what was
		//       actually written. Should we change this?
		throughput := (float64(cfg.Size) / 1000 / 1000) / (float64(avgTook) / float64(time.Second))
		fn(Result{
			Name:        fmt.Sprintf("%s:%s_%s", cfg.BenchName, cfg.InputName, hint),
			Encryption:  string(hint.EncryptionAlgo),
			Compression: string(hint.CompressionAlgo),
			Config:      cfg,
			Took:        avgTook,
			Throughput:  throughput,
		})

		if !supportsHints {
			// If there are no hints there is no point.
			// of repeating the benchmark several times.
			break
		}
	}

	return nil
}

// IPFS is expensive to set-up, so let's do it only once.
func ipfsIsNeeded(cfgs []Config) bool {
	for _, cfg := range cfgs {
		if strings.Contains(strings.ToLower(cfg.BenchName), "ipfs") {
			return true
		}
	}

	return false
}

// Benchmark runs the benchmarks specified by `cfgs` and call `fn` on each result.
func Benchmark(cfgs []Config, fn func(result Result)) error {
	needsIPFS := ipfsIsNeeded(cfgs)
	var result *setup.Result

	if needsIPFS {
		var err error
		log.Infof("Setting up IPFS for the benchmarks...")

		ipfsPath, err := ioutil.TempDir("", "brig-iobench-ipfs-repo-*")
		if err != nil {
			return err
		}

		result, err = setup.IPFS(setup.Options{
			LogWriter:        ioutil.Discard,
			Setup:            true,
			SetDefaultConfig: true,
			SetExtraConfig:   true,
			IpfsPath:         ipfsPath,
			InitProfile:      "test",
		})

		if err != nil {
			return err
		}
	}

	for _, cfg := range cfgs {
		var ipfsPath string
		if result != nil {
			ipfsPath = result.IpfsPath
		}

		if err := benchmarkSingle(cfg, fn, ipfsPath); err != nil {
			return err
		}
	}

	if needsIPFS {
		if result.IpfsPath != "" {
			os.RemoveAll(result.IpfsPath)
		}

		if result.PID > 0 {
			proc, err := os.FindProcess(result.PID)
			if err != nil {
				log.WithError(err).Warnf("failed to get IPFS PID")
			} else {
				if err := proc.Kill(); err != nil {
					log.WithError(err).Warnf("failed to kill IPFS PID")
				}
			}
		}
	}

	return nil
}
