package pwutil

import (
	"context"
	"os/exec"
	"os/user"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// ReadPasswordFromHelper tries to read a password from a shell command.
// The shell command gets BRIG_PATH and ENV as environment variables.
// The output of the password is trimmed from newlines.
func ReadPasswordFromHelper(basePath, helperCommand string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", helperCommand) // #nosec
	cmd.Env = append(cmd.Env, "BRIG_PATH="+basePath)
	cmd.Env = append(cmd.Env, "HOME="+currentUser.HomeDir)

	data, err := cmd.Output()
	if err != nil {
		log.Warningf("failed to execute password helper: %v: %s", err, data)
		return "", err
	}

	return strings.Trim(string(data), "\n"), nil
}
