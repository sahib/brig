package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// indodeString convert file path a hardware dependent string in the form DeviceID/Inode
func inodeString(path string) (string, error) {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return path, err
	}
	s := fmt.Sprintf("%d/%d", stat.Dev, stat.Ino)
	return s, nil
}
