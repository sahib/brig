package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

// Encoder defines how the config can be serialized to a byte stream
type Encoder interface {
	// Encode takes the version and the internal config representation
	// and turns it to a byte stream that was passed when creatin the encoder.
	Encode(version Version, data map[interface{}]interface{}) error
}

// Decoder defines how a byte stream can be parsed to a config
type Decoder interface {
	// Decode takes a byte stream from when the decoder was created
	// and parses the version and the internal representation out of it.
	Decode() (Version, map[interface{}]interface{}, error)
}

////////////

type yamlEncoder struct {
	w io.Writer
}

// NewYamlEncoder creates a new Encoder that writes a YAML file with
// the config data. The file will start with a comment indicating the version,
// so pay attention to not remove it by accident.
func NewYamlEncoder(w io.Writer) Encoder {
	return &yamlEncoder{w: w}
}

func (ye *yamlEncoder) Encode(version Version, memory map[interface{}]interface{}) error {
	data, err := yaml.Marshal(memory)
	if err != nil {
		return err
	}

	// Build the version header:
	header := []byte(fmt.Sprintf(
		"# version: %d (DO NOT MODIFY THIS LINE)\n",
		version,
	))

	// Write the prefixed yaml data:
	_, err = ye.w.Write(append(header, data...))
	return err
}

////////////

type yamlDecoder struct {
	r io.Reader
}

// NewYamlDecoder creates a new Decoder that parses the data in `r`.
// It will look at the first line of the input to get the version.
func NewYamlDecoder(r io.Reader) Decoder {
	return &yamlDecoder{r: r}
}

func readVersionFromData(data []byte) (Version, error) {
	match := versionTag.FindSubmatch(data)
	if match == nil {
		return 0, ErrNotVersioned
	}

	if len(match) < 2 {
		return 0, ErrNotVersioned
	}

	version, err := strconv.ParseInt(string(match[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	return Version(version), nil
}

func (yd *yamlDecoder) Decode() (Version, map[interface{}]interface{}, error) {
	data, err := ioutil.ReadAll(yd.r)
	if err != nil {
		return Version(-1), nil, err
	}

	version, err := readVersionFromData(data)
	if err != nil && err != ErrNotVersioned {
		return Version(-1), nil, err
	}

	memory := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, memory); err != nil {
		return Version(-1), nil, err
	}

	return version, memory, nil
}
