package hints

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/klauspost/cpuid/v2"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util/trie"
	"github.com/sahib/config"
)

var (
	// ErrNoSuchHint is returned by Remove when there is no hint at this path.
	ErrNoSuchHint = errors.New("no such hint at this path")

	// ErrInvalidHint is returned upon setting an invalid hint.
	ErrInvalidHint = errors.New("invalid hint")
)

// CompressionHint is an enumeration of possible compression types.
type CompressionHint string

const (
	// CompressionNone leaves the stream as-is.
	CompressionNone = CompressionHint("none")

	// CompressionLZ4 compresses the stream in lz4 mode.
	CompressionLZ4 = CompressionHint("lz4")

	// CompressionSnappy  compresses the stream in snappy mode.
	CompressionSnappy = CompressionHint("snappy")

	// CompressionGuess tries to guess a suitable type by looking at
	// different aspects of the stream.
	CompressionGuess = CompressionHint("guess")
)

var (
	compressionHintMap = map[CompressionHint]compress.AlgorithmType{
		CompressionNone:   compress.AlgoUnknown,
		CompressionLZ4:    compress.AlgoLZ4,
		CompressionSnappy: compress.AlgoSnappy,
		CompressionGuess:  compress.AlgoUnknown,
	}

	compressionSortMap = map[CompressionHint]int{
		CompressionNone:   0,
		CompressionLZ4:    1,
		CompressionSnappy: 2,
		CompressionGuess:  3,
	}
)

// IsValid returns true if `ch` is a valid compression hint.
func (ch CompressionHint) IsValid() bool {
	_, ok := compressionHintMap[ch]
	return ok
}

// ToCompressAlgorithmType converts the hint to the enum used in compress
func (ch CompressionHint) ToCompressAlgorithmType() compress.AlgorithmType {
	return compressionHintMap[ch]
}

// CompressAlgorithmTypeToCompressionHint is a very aptly named function
// that converts `algo` to a hint. This is not a perfect conversion, since
// compress package doesn't know any "none" or "guess" algorithm.
func CompressAlgorithmTypeToCompressionHint(algo compress.AlgorithmType) CompressionHint {
	switch algo {
	case compress.AlgoUnknown:
		return CompressionNone
	case compress.AlgoLZ4:
		return CompressionLZ4
	case compress.AlgoSnappy:
		return CompressionSnappy
	default:
		return CompressionNone
	}
}

// ValidCompressionHints returns all valid compression hints.
func ValidCompressionHints() []string {
	s := []string{}
	for h := range compressionHintMap {
		s = append(s, string(h))
	}

	return s
}

// CompressionHints returns all possible compression hints.
func CompressionHints() []CompressionHint {
	s := []CompressionHint{}

	for compressionHint := range compressionHintMap {
		s = append(s, compressionHint)
	}

	return s
}

// EncryptionHint is an enum of valid encryption types
type EncryptionHint string

const (
	// EncryptionNone disables all encryption on the stream.
	EncryptionNone = EncryptionHint("none")

	// EncryptionAES256GCM uses AES256 in GCM mode.
	EncryptionAES256GCM = EncryptionHint("aes256gcm")

	// EncryptionChaCha20 uses ChaCha20 with Poly1305 as MAC.
	EncryptionChaCha20 = EncryptionHint("chacha20")
)

var (
	encryptionHintMap = map[EncryptionHint]encrypt.Flags{
		EncryptionNone:      encrypt.FlagEmpty,
		EncryptionAES256GCM: encrypt.FlagEncryptAES256GCM,
		EncryptionChaCha20:  encrypt.FlagEncryptChaCha20,
	}

	encryptionSortMap = map[EncryptionHint]int{
		EncryptionNone:      0,
		EncryptionAES256GCM: 1,
		EncryptionChaCha20:  2,
	}
)

// IsValid checks if `eh` is a valid encryption type
func (eh EncryptionHint) IsValid() bool {
	_, ok := encryptionHintMap[eh]
	return ok
}

// ToEncryptFlags returns flags suitable for passing to the encrypt.NewWriter.
func (eh EncryptionHint) ToEncryptFlags() encrypt.Flags {
	return encryptionHintMap[eh]
}

// ValidEncryptionHints returns all valid encryption hints.
func ValidEncryptionHints() []string {
	s := []string{}
	for h := range encryptionHintMap {
		s = append(s, string(h))
	}

	return s
}

// Hint describes the settings brig applies to streams.
type Hint struct {
	// CompressionAlgo can be an algorithm or "guess"
	// to let brig choose a suitable one.
	CompressionAlgo CompressionHint

	// EncryptionAlgo must be a valid encryption algorithm.
	EncryptionAlgo EncryptionHint
}

// Small heuristic to decide if we should use ChaCha20
// or AES for encryption as default.
var (
	cpuInfoOnce   sync.Once
	cpuHasNoAESNI int32
)

// Default returns the default stream settings
func Default() Hint {
	cpuInfoOnce.Do(func() {
		if !cpuid.CPU.Supports(cpuid.AESNI) {
			atomic.StoreInt32(&cpuHasNoAESNI, 1)
		}
	})

	encHint := EncryptionAES256GCM
	if atomic.LoadInt32(&cpuHasNoAESNI) > 0 {
		encHint = EncryptionChaCha20
	}

	return Hint{
		EncryptionAlgo:  encHint,
		CompressionAlgo: CompressionGuess,
	}
}

// IsValid checks if all fields of the hint are valid.
func (h Hint) IsValid() bool {
	return h.EncryptionAlgo.IsValid() && h.CompressionAlgo.IsValid()
}

// EncryptFlags returns combined flags for encrypt.NewWriter.
// If valid compression is set, then FlagCompressedInside is OR'd in.
func (h Hint) EncryptFlags() encrypt.Flags {
	flags := h.EncryptionAlgo.ToEncryptFlags()
	if h.CompressionAlgo != CompressionNone {
		flags |= encrypt.FlagCompressedInside
	}

	return flags
}

// IsRaw checks if the stream can be read directly from IPFS.
func (h Hint) IsRaw() bool {
	return h.EncryptionAlgo == EncryptionNone && h.CompressionAlgo == CompressionNone
}

func (h Hint) String() string {
	return fmt.Sprintf("enc:%s-zip:%s", h.EncryptionAlgo, h.CompressionAlgo)
}

// Less returns false if `o` should be sorted before `h`.
func (h Hint) Less(o Hint) bool {
	// This sorts "none" always before any other hint type.
	// Leverages the fact that the enum value of none is lower than all others.
	encNumH, ok := encryptionSortMap[h.EncryptionAlgo]
	if !ok {
		encNumH = int(^uint(0) >> 1)
	}

	encNumO, ok := encryptionSortMap[o.EncryptionAlgo]
	if !ok {
		encNumO = int(^uint(0) >> 1)
	}

	if encNumH != encNumO {
		return encNumH < encNumO
	}

	zipNumH, ok := compressionSortMap[h.CompressionAlgo]
	if !ok {
		zipNumH = int(^uint(0) >> 1)
	}

	zipNumO, ok := compressionSortMap[o.CompressionAlgo]
	if !ok {
		zipNumO = int(^uint(0) >> 1)
	}

	return zipNumH < zipNumO
}

// EncryptionHints returns all possible encryption hints.
func EncryptionHints() []EncryptionHint {
	s := []EncryptionHint{}

	for encryptionHint := range encryptionHintMap {
		s = append(s, encryptionHint)
	}

	return s
}

// AllPossibleHints returns all possible valid hint combination.
// Useful for testing, but might be useful for cmdline purposes too.
func AllPossibleHints() []Hint {
	hints := []Hint{}

	for compressionHint := range compressionHintMap {
		for encryptionHint := range encryptionHintMap {
			hints = append(hints, Hint{
				CompressionAlgo: compressionHint,
				EncryptionAlgo:  encryptionHint,
			})
		}
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Less(hints[j])
	})

	return hints
}

var (
	defaults = config.DefaultMapping{
		"hints": config.DefaultMapping{
			"__many__": config.DefaultMapping{
				"path": config.DefaultEntry{
					Default:      "",
					NeedsRestart: false,
					Docs:         "The path to apply the hints to. Recursive if directory.",
				},
				"compression_algo": config.DefaultEntry{
					Default:      string(Default().CompressionAlgo),
					NeedsRestart: false,
					Docs:         "Which compression algorithm to use.",
					Validator:    config.EnumValidator(ValidCompressionHints()...),
				},
				"encryption_algo": config.DefaultEntry{
					Default:      string(Default().EncryptionAlgo),
					NeedsRestart: false,
					Docs:         "Which encryption algorithm to use.",
					Validator:    config.EnumValidator(ValidEncryptionHints()...),
				},
			},
		},
	}
)

func prefixSlash(path string) string {
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}

	return path
}

// HintManager is a helper to store hints for certain paths.
type HintManager struct {
	mu   sync.Mutex
	root *trie.Node
}

// NewManager reads a YAML file from `yamlReader`.
// If the reader is nil, then an empty file is assumed.
// There is always a root hint with the settings returned by Default()
//
// All methods are safe to call from several go routines.
func NewManager(yamlReader io.Reader) (*HintManager, error) {
	if yamlReader == nil {
		// If no hint manager was loaded, then let's load one
		// that always returns the defaults.
		return &HintManager{
			root: trie.NewNodeWithData(Default()),
		}, nil
	}

	mgr := config.NewMigrater(1, config.StrictnessWarn)
	mgr.Add(0, nil, defaults)

	cfg, err := mgr.Migrate(config.NewYamlDecoder(yamlReader))
	if err != nil {
		return nil, e.Wrap(err, "failed to migrate or open")
	}

	root := trie.NewNode()

	hintMapping := cfg.Section("hints")
	for _, key := range hintMapping.Keys() {
		if !strings.HasSuffix(key, ".path") {
			continue
		}

		hintPath := hintMapping.String(key)
		prefixKey := strings.TrimSuffix(key, ".path")

		hint := Hint{
			CompressionAlgo: CompressionHint(hintMapping.String(prefixKey + ".compression_algo")),
			EncryptionAlgo:  EncryptionHint(hintMapping.String(prefixKey + ".encryption_algo")),
		}

		// Fill up a trie with each hint:
		root.InsertWithData(prefixSlash(hintPath), hint)
	}

	return &HintManager{
		root: root,
	}, nil
}

// Lookup will give a hint for path. If there is no such hint,
// we return the default. If we don't have a hint for `path` directly
// the hint of the nearest parent is returned. If that also did not
// work (for whatever reason) then the default is returned.
// The returned hint is valid in any case.
func (hm *HintManager) Lookup(path string) Hint {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	path = prefixSlash(path)
	node := hm.root.LookupDeepest(path)
	if node == nil || node.Data == nil {
		// This can happen only if the root node
		// does not have any data.
		return Default()
	}

	return node.Data.(Hint)
}

// Set remembers a `hint` for `path`.
func (hm *HintManager) Set(path string, hint Hint) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	path = prefixSlash(path)
	if !hint.IsValid() {
		return ErrInvalidHint
	}

	hm.root.InsertWithData(path, hint)
	return nil
}

// Remove forgets a hint at `path`.
func (hm *HintManager) Remove(path string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	path = prefixSlash(path)
	nd := hm.root.Lookup(path)
	if nd == nil || nd.Data == nil {
		return ErrNoSuchHint
	}

	nd.Remove()
	return nil
}

// List returns a map of all paths with their corresponding hints.
func (hm *HintManager) List() map[string]Hint {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	return hm.list()
}

// list() is used both by Save() and List()
func (hm *HintManager) list() map[string]Hint {
	hints := make(map[string]Hint)

	hm.root.Walk(true, func(node *trie.Node) bool {
		if node.Data == nil {
			return true
		}

		path := prefixSlash(node.Path())
		hints[path] = node.Data.(Hint)
		return true
	})

	return hints
}

// Save writes a YAML representation of the hints to `w`.
func (hm *HintManager) Save(w io.Writer) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	emptyCfg, err := config.Open(nil, defaults, config.StrictnessWarn)
	if err != nil {
		return err
	}

	hintMapping := emptyCfg.Section("hints")
	for path, hint := range hm.list() {
		hintMapping.SetString(path+".path", path)
		hintMapping.SetString(path+".compression_algo", string(hint.CompressionAlgo))
		hintMapping.SetString(path+".encryption_algo", string(hint.EncryptionAlgo))
	}

	return emptyCfg.Save(config.NewYamlEncoder(w))
}
