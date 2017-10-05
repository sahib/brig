package client

import "github.com/disorganizer/brig/brigd/capnp"

func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Meta_ping_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return err
	}

	_, err = result.Reply()
	return err
}

// Quit sends a quit signal to brigd.
func (cl *Client) Quit() error {
	call := cl.api.Quit(cl.ctx, func(p capnp.Meta_quit_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Init(path, owner, backend string) error {
	call := cl.api.Init(cl.ctx, func(p capnp.Meta_init_Params) error {
		if err := p.SetOwner(owner); err != nil {
			return err
		}

		if err := p.SetBasePath(path); err != nil {
			return err
		}

		return p.SetBackend(backend)
	})

	_, err := call.Struct()
	return err
}
