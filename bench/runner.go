package bench

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/repo/setup"
)

// TODO:
// Output sysinfo at front:
// CPU-model, num cores.

type Config struct {
	InputName   string `json:"input_name"`
	BenchName   string `json:"bench_name"`
	Size        uint64 `json:"size"`
	Encryption  string
	Compression string
}

type Result struct {
	Config      Config        `json:"config"`
	Encryption  string        `json:"encryption"`
	Compression string        `json:"compression"`
	Took        time.Duration `json:"took"`
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

	out, err := BenchByName(cfg.BenchName, ipfsPath)
	if err != nil {
		return err
	}

	defer out.Close()

	for _, hint := range sortHints(buildHints(cfg)) {
		supportsHints := out.SupportHints()
		if !supportsHints {
			// Indicate in output that nothing was encrypted or compressed.
			hint.CompressionAlgo = hints.CompressionNone
			hint.EncryptionAlgo = hint.EncryptionAlgo
		}

		r, err := in.Reader()
		if err != nil {
			return err
		}

		took, err := out.Bench(hint, r)
		if err != nil {
			return err
		}

		fn(Result{
			Encryption:  string(hint.EncryptionAlgo),
			Compression: string(hint.CompressionAlgo),
			Config:      cfg,
			Took:        took,
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

func Benchmark(cfgs []Config, fn func(result Result)) error {
	needsIPFS := ipfsIsNeeded(cfgs)

	var (
		ipfsPath string
		ipfsPID  int
	)

	if needsIPFS {
		var err error
		log.Warnf("Setting up IPFS for the benchmarks...")

		ipfsPath, err = ioutil.TempDir("", "brig-iobench-ipfs-repo-*")
		if err != nil {
			return err
		}

		_, ipfsPID, err = setup.IPFS(ioutil.Discard, true, true, true, ipfsPath)
		if err != nil {
			return err
		}
	}

	for _, cfg := range cfgs {
		if err := benchmarkSingle(cfg, fn, ipfsPath); err != nil {
			return err
		}
	}

	if needsIPFS {
		if ipfsPath != "" {
			os.RemoveAll(ipfsPath)
		}

		if ipfsPID > 0 {
			proc, err := os.FindProcess(ipfsPID)
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
