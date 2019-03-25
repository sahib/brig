// +build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

var Aliases = map[string]interface{}{
	"b": Build.Binary,
	"g": Build.Generate,
	"t": Build.Test,
	"l": Dev.Lint,
	"c": Dev.Capnp,
}

/////////////////////
// UTILITY HELPERS //
/////////////////////

func speak(format string, args ...interface{}) {
	if mg.Verbose() {
		fmt.Printf("-- "+format+"\n", args...)
	}
}

func readVersion() (*semver.Version, error) {
	data, err := ioutil.ReadFile(".version")
	if err != nil {
		return nil, fmt.Errorf("failed to read .version file: %v", err)
	}

	if bytes.HasPrefix(data, []byte{'v'}) {
		data = data[1:]
	}

	data = bytes.TrimSpace(data)
	vers, err := semver.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse .version: %v", err)
	}

	return &vers, nil
}

func gitRev() string {
	rev, err := sh.Output("git", "rev-parse", "HEAD")
	if err != nil {
		speak("-- could not get git version: %v", err)
		return ""
	}

	return rev
}

func binaryOutput() string {
	path := os.Getenv("BRIG_BINARY_PATH")
	if path != "" {
		speak("using binary path from BRIG_BINARY_PATH: %s", path)
		return path
	}

	speak("using ${GOBIN}/brig as binary output location")
	return filepath.Join(os.Getenv("GOBIN"), "brig")
}

////////////////////
// ACTUAL TARGETS //
////////////////////

var Default = Build.Binary

type Build mg.Namespace

func (Build) Generate() error {
	// Check if we really need to to do the rather
	// expensive "go generate" step.
	modified, err := target.Dir(
		// Reference file:
		"gateway/static/resource.go",
		// Paths to be checked for their age:
		"gateway/static/favicon.ico",
		"gateway/static/webfonts",
		"gateway/static/css",
		"gateway/static/js",
	)

	if err != nil {
		return err
	}

	if !modified {
		speak("ignoring generate; source did not change.")
		return nil
	}

	return sh.Run("go", "generate", "./...")
}

func (Build) Binary() error {
	mg.Deps(Build.Generate)

	version, err := readVersion()
	if err != nil {
		return err
	}

	releaseType := ""
	if len(version.Pre) > 0 {
		releaseType = version.Pre[0].String()
	}

	imp := "github.com/sahib/brig/version"
	ldflags := []string{
		"-X", fmt.Sprintf("%s.Major=%d", imp, version.Major),
		"-X", fmt.Sprintf("%s.Minor=%d", imp, version.Minor),
		"-X", fmt.Sprintf("%s.Patch=%d", imp, version.Patch),
		"-X", fmt.Sprintf("%s.ReleaseType=%s", imp, releaseType),
		"-X", fmt.Sprintf("%s.BuildTime=%s", imp, time.Now().Format(time.RFC3339)),
		"-X", fmt.Sprintf("%s.GitRev=%s", imp, gitRev()),
	}

	useUPX := false
	switch os.Getenv("BRIG_SMALL_BINARY") {
	case "tiny":
		useUPX = true
		fallthrough
	case "small":
		ldflags = append(ldflags, "-s")
		ldflags = append(ldflags, "-w")
	default:
		break
	}

	binPath := binaryOutput()
	minusld := strings.Join(ldflags, " ")
	err = sh.Run("go", "build", "-ldflags", minusld, "-o", binPath)
	if err != nil {
		return err
	}

	if useUPX {
		if err := sh.Run("upx", binPath); err != nil {
			return err
		}
	}

	return nil
}

func (Build) Test() error {
	mg.Deps(Build.Generate)

	return sh.RunV("go", "test", "./...")
}

// Development tools that are not relevant to the user's building process:
type Dev mg.Namespace

func (Dev) Capnp() error {
	capnp := func(path string) error {
		return sh.Run(
			"capnp", "compile",
			"-I/home/sahib/go/src/zombiezen.com/go/capnproto2/std",
			"-ogo", path,
		)
	}

	if err := capnp("server/capnp/local_api.capnp"); err != nil {
		return err
	}

	if err := capnp("catfs/nodes/capnp/nodes.capnp"); err != nil {
		return err
	}

	if err := capnp("net/capnp/api.capnp"); err != nil {
		return err
	}

	if err := capnp("catfs/vcs/capnp/patch.capnp"); err != nil {
		return err
	}

	if err := capnp("catfs/capnp/pinner.capnp"); err != nil {
		return err
	}

	if err := capnp("events/capnp/events_api.capnp"); err != nil {
		return err
	}

	if err := capnp("gateway/db/capnp/user.capnp"); err != nil {
		return err
	}

	return nil
}

func (Dev) Lint() error {
	findCmd := "find -iname '*.go' -type f ! -path '*vendor*' ! -path '*capnp*' ! -iname 'build.go'"

	linters := []string{
		fmt.Sprintf("%s -exec gofmt -s -w {} \\;", findCmd),
		fmt.Sprintf("%s -exec go fix {} \\;", findCmd),
		fmt.Sprintf("%s -exec golint {} \\;", findCmd),
		fmt.Sprintf("%s -exec misspell {} \\;", findCmd),
		fmt.Sprintf("%s -exec gocyclo -over 20 {} \\; | sort -n", findCmd),
	}

	for _, linter := range linters {
		sh.RunV("sh", "-c", linter)
	}

	return nil
}

func (Dev) Cloc() error {
	cmd := "cloc $(find -iname '*.elm' -or -iname '*.go' -a ! -path '*vendor*' ! -path '*capnp*' | head -n -1 | sort | uniq)"
	return sh.RunV("sh", "-c", cmd)
}
