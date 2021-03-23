// +build windows

package cmd

// indodeString convert file path a hardware dependent string
// unfortunately on non unix platforms DeviceID and Inode are unavailable
// so we return back the file path
func inodeString(path string) (string, error) {
	return path, nil
}
