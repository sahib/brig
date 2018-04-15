package cmds

import (
	"bufio"
	"context"
	"fmt"
	"reflect"

	"gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"
	"gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"
)

// Request represents a call to a command from a consumer
type Request struct {
	Context       context.Context
	Root, Command *Command

	Path      []string
	Arguments []string
	Options   cmdkit.OptMap

	Files files.File
}

// NewRequest returns a request initialized with given arguments
// An non-nil error will be returned if the provided option values are invalid
func NewRequest(ctx context.Context, path []string, opts cmdkit.OptMap, args []string, file files.File, root *Command) (*Request, error) {
	if opts == nil {
		opts = make(cmdkit.OptMap)
	}

	cmd, err := root.Get(path)
	if err != nil {
		return nil, err
	}

	req := &Request{
		Path:      path,
		Options:   opts,
		Arguments: args,
		Files:     file,
		Root:      root,
		Command:   cmd,
		Context:   ctx,
	}

	return req, req.convertOptions(root)
}

type allArgsCovered struct{}

func (allArgsCovered) Error() string            { return "all arguments covered by positional arguments" }
func (allArgsCovered) ArgsAlreadyCovered() bool { return true }

type moreArgsExpected struct{}

func (moreArgsExpected) Error() string          { return "expected more arguments from stdin" }
func (moreArgsExpected) MoreArgsExpected() bool { return true }

// BodyArgs returns a scanner that returns arguments passed in the body as tokens.
func (req *Request) BodyArgs() (*bufio.Scanner, error) {
	if len(req.Arguments) >= len(req.Command.Arguments) {
		return nil, allArgsCovered{}
	}

	if req.Files == nil {
		return nil, moreArgsExpected{}
	}

	fi, err := req.Files.NextFile()
	if err != nil {
		return nil, err
	}

	return bufio.NewScanner(fi), nil
}

type argsAlreadyCovereder interface {
	ArgsAlreadyCovered() bool
}

func IsAllArgsAlreadyCovered(err error) bool {
	argsErr, ok := err.(argsAlreadyCovereder)
	return ok && argsErr.ArgsAlreadyCovered()
}

type moreArgsExpecteder interface {
	MoreArgsExpected() bool
}

func IsMoreArgumentsExpected(err error) bool {
	argsErr, ok := err.(moreArgsExpecteder)
	return ok && argsErr.MoreArgsExpected()
}

func (req *Request) ParseBodyArgs() error {
	s, err := req.BodyArgs()
	if err != nil {
		return err
	}

	for s.Scan() {
		req.Arguments = append(req.Arguments, s.Text())
	}

	return s.Err()
}

func (req *Request) SetOption(name string, value interface{}) {
	optDefs, err := req.Root.GetOptions(req.Path)
	optDef, found := optDefs[name]

	if req.Options == nil {
		req.Options = map[string]interface{}{}
	}

	// unknown option, simply set the value and return
	// TODO we might error out here instead
	if err != nil || !found {
		req.Options[name] = value
		return
	}

	name = optDef.Name()
	req.Options[name] = value

	return
}

func (req *Request) convertOptions(root *Command) error {
	optDefs, err := root.GetOptions(req.Path)
	if err != nil {
		return err
	}

	for k, v := range req.Options {
		opt, ok := optDefs[k]
		if !ok {
			continue
		}

		kind := reflect.TypeOf(v).Kind()
		if kind != opt.Type() {
			if str, ok := v.(string); ok {
				val, err := opt.Parse(str)
				if err != nil {
					value := fmt.Sprintf("value %q", v)
					if len(str) == 0 {
						value = "empty value"
					}
					return fmt.Errorf("Could not convert %q to type %q (for option %q)",
						value, opt.Type().String(), "-"+k)
				}
				req.Options[k] = val

			} else {
				return fmt.Errorf("Option %q should be type %q, but got type %q",
					k, opt.Type().String(), kind.String())
			}
		}

		for _, name := range opt.Names() {
			if _, ok := req.Options[name]; name != k && ok {
				return fmt.Errorf("Duplicate command options were provided (%q and %q)",
					k, name)
			}
		}
	}

	return nil
}

// GetEncoding returns the EncodingType set in a request, falling back to JSON
func GetEncoding(req *Request) EncodingType {
	encIface := req.Options[EncLong]
	if encIface == nil {
		return JSON
	}

	switch enc := encIface.(type) {
	case string:
		return EncodingType(enc)
	case EncodingType:
		return enc
	default:
		return JSON
	}
}

// fillDefault fills in default values if option has not been set
func (req *Request) FillDefaults() error {
	optDefMap, err := req.Root.GetOptions(req.Path)
	if err != nil {
		return err
	}

	optDefs := map[cmdkit.Option]struct{}{}

	for _, optDef := range optDefMap {
		optDefs[optDef] = struct{}{}
	}

Outer:
	for optDef := range optDefs {
		dflt := optDef.Default()
		if dflt == nil {
			// option has no dflt, continue
			continue
		}

		names := optDef.Names()
		for _, name := range names {
			if _, ok := req.Options[name]; ok {
				// option has been set, continue with next option
				continue Outer
			}
		}

		req.Options[optDef.Name()] = dflt
	}

	return nil
}
