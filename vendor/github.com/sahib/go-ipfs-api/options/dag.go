package options

// DagPutSettings is a set of DagPut options.
type DagPutSettings struct {
	InputEnc string
	Kind     string
	Pin      string
}

// DagPutOption is a single DagPut option.
type DagPutOption func(opts *DagPutSettings) error

// DagPutOptions applies the given options to a DagPutSettings instance.
func DagPutOptions(opts ...DagPutOption) (*DagPutSettings, error) {
	options := &DagPutSettings{
		InputEnc: "json",
		Kind:     "cbor",
		Pin:      "false",
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}
	return options, nil
}

type dagOpts struct{}

var Dag dagOpts

// Pin is an option for Dag.Put which specifies whether to pin the added
// dags. Default is "false".
func (dagOpts) Pin(pin string) DagPutOption {
	return func(opts *DagPutSettings) error {
		opts.Pin = pin
		return nil
	}
}

// InputEnc is an option for Dag.Put which specifies the input encoding of the
// data. Default is "json", most formats/codecs support "raw".
func (dagOpts) InputEnc(enc string) DagPutOption {
	return func(opts *DagPutSettings) error {
		opts.InputEnc = enc
		return nil
	}
}

// Kind is an option for Dag.Put which specifies the format that the dag
// will be added as. Default is "cbor".
func (dagOpts) Kind(kind string) DagPutOption {
	return func(opts *DagPutSettings) error {
		opts.Kind = kind
		return nil
	}
}
