package hints

import (
	"io"
	"strings"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/util/trie"
	"github.com/sahib/config"
)

type nothing struct{}

type CompressionHint string

const (
	CompressionNone   = "none"
	CompressionLZ4    = "lz4"
	CompressionSnappy = "snappy"
	CompressionGuess  = "guess"
)

var (
	compressionHintMap = map[CompressionHint]nothing{
		CompressionNone:   nothing{},
		CompressionLZ4:    nothing{},
		CompressionSnappy: nothing{},
		CompressionGuess:  nothing{},
	}
)

func (ch CompressionHint) IsValid() bool {
	_, ok := compressionHintMap[ch]
	return ok
}

func validCompressionHints() []string {
	s := []string{}
	for h := range compressionHintMap {
		s = append(s, string(h))
	}

	return s
}

type EncryptionHint string

const (
	EncryptionNone      = "none"
	EncryptionAES256GCM = "aes256gcm"
	EncryptionChaCha20  = "chacha20"
)

var (
	encryptionHintMap = map[EncryptionHint]nothing{
		EncryptionNone:      nothing{},
		EncryptionAES256GCM: nothing{},
		EncryptionChaCha20:  nothing{},
	}
)

func (eh EncryptionHint) IsValid() bool {
	_, ok := encryptionHintMap[eh]
	return ok
}

func validEncryptionHints() []string {
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

func Default() Hint {
	return Hint{
		EncryptionAlgo:  EncryptionAES256GCM,
		CompressionAlgo: CompressionGuess,
	}
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
					Default:      CompressionGuess,
					NeedsRestart: false,
					Docs:         "Which compression algorithm to use.",
					Validator:    config.EnumValidator(validCompressionHints()...),
				},
				"encryption_algo": config.DefaultEntry{
					Default:      "guess",
					NeedsRestart: false,
					Docs:         "Which encryption algorithm to use.",
					Validator:    config.EnumValidator(validEncryptionHints()...),
				},
			},
		},
	}
)

type HintManager struct {
	root *trie.Node
}

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
		root.InsertWithData(hintPath, hint)
	}

	return &HintManager{
		root: root,
	}, nil
}

func (hm *HintManager) Lookup(path string) Hint {
	node := hm.root.LookupDeepest(path)
	if node == nil {
		// This can happen only if the root node
		// does not have any data.
		return Default()
	}

	return node.Data.(Hint)
}

func (hm *HintManager) Remember(path string, hint Hint) {
	hm.root.InsertWithData(path, hint)
}

func (hm *HintManager) Save(w io.Writer) error {
	emptyCfg, err := config.Open(nil, defaults, config.StrictnessWarn)
	if err != nil {
		return err
	}

	hintMapping := emptyCfg.Section("hints")

	hm.root.Walk(true, func(node *trie.Node) bool {
		if node.Data == nil {
			return true
		}

		hint := node.Data.(Hint)
		path := node.Path()
		hintMapping.SetString(path+".path", path)
		hintMapping.SetString(path+".compression_algo", string(hint.CompressionAlgo))
		hintMapping.SetString(path+".encryption_algo", string(hint.EncryptionAlgo))
		return true
	})

	return emptyCfg.Save(config.NewYamlEncoder(w))
}
