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

func (cl *Client) ConfigGet(key string) (string, error) {
	call := cl.api.ConfigGet(cl.ctx, func(p capnp.Meta_configGet_Params) error {
		return p.SetKey(key)
	})

	result, err := call.Struct()
	if err != nil {
		return "", err
	}

	return result.Value()
}

func (cl *Client) ConfigSet(key, value string) error {
	call := cl.api.ConfigSet(cl.ctx, func(p capnp.Meta_configSet_Params) error {
		if err := p.SetValue(value); err != nil {
			return err
		}

		return p.SetKey(key)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) ConfigAll() (map[string]string, error) {
	call := cl.api.ConfigAll(cl.ctx, func(p capnp.Meta_configAll_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	pairs, err := result.All()
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]string)

	for idx := 0; idx < pairs.Len(); idx++ {
		pair := pairs.At(idx)
		key, err := pair.Key()
		if err != nil {
			return nil, err
		}

		val, err := pair.Val()
		if err != nil {
			return nil, err
		}

		configMap[key] = val
	}

	return configMap, nil
}
