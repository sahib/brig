// +build !linux

package server

func increaseMaxOpenFds() error {
	// no op on for non-linux systems.
	return nil
}
