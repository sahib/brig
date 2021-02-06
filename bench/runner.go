package bench

import (
	"sort"
	"time"

	"github.com/sahib/brig/repo/hints"
)

type Config struct {
	InputName   string `json:"input_name"`
	BenchName   string `json:"bench_name"`
	Size        uint64 `json:"size"`
	Random      bool   `json:"random"`
	Encryption  string
	Compression string
}

type Result struct {
	Config      Config        `json:"config"`
	Encryption  string        `json:"encryption"`
	Compression string        `json:"compression"`
	Took        time.Duration `json:"took"`
}

func withTiming(fn func() error) (time.Duration, error) {
	start := time.Now()
	if err := fn(); err != nil {
		return time.Since(start), err
	}

	return time.Since(start), nil
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

func benchmarkSingle(cfg Config, fn func(result Result)) error {
	in, err := InputByName(cfg.InputName, cfg.Size, cfg.Random)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := BenchByName(cfg.BenchName)
	if err != nil {
		return err
	}

	defer out.Close()

	for _, hint := range sortHints(buildHints(cfg)) {
		supportsHints := out.SupportHints()
		if supportsHints {
			if err := out.SetHint(hint); err != nil {
				return err
			}
		} else {
			// Indicate in output that nothing was encrypted or compressed.
			hint.CompressionAlgo = hints.CompressionNone
			hint.EncryptionAlgo = hint.EncryptionAlgo
		}

		r, err := in.Reader()
		if err != nil {
			return err
		}

		took, err := withTiming(func() error {
			return out.Process(r)
		})

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

func Benchmark(cfgs []Config, fn func(result Result)) error {
	for _, cfg := range cfgs {
		if err := benchmarkSingle(cfg, fn); err != nil {
			return err
		}
	}

	return nil
}
