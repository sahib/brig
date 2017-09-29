package client

import (
	"fmt"

	"github.com/disorganizer/brig/brigd/capnp"
)

func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Meta_ping_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		fmt.Println("Add #1 failed:", err)
		return err
	}

	_, err = result.Reply()
	return err
}
