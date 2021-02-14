package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/mr-tron/base58"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/fuse/fusetest"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/util/testutil"
	"github.com/urfave/cli"
)

func handleDebugPprofPort(ctx *cli.Context, ctl *client.Client) error {
	port, err := ctl.DebugProfilePort()
	if err != nil {
		return err
	}

	if port > 0 {
		fmt.Println(port)
	} else {
		fmt.Println("Profiling is not enabled.")
		fmt.Println("Enable daemon.enable_pprof and restart.")
	}

	return nil
}

func readDebugKey(ctx *cli.Context) ([]byte, error) {
	keyB58 := ctx.String("key")
	key, err := base58.Decode(keyB58)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func handleDebugDecodeStream(ctx *cli.Context) error {
	key, err := readDebugKey(ctx)
	if err != nil {
		return err
	}

	fd, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	defer fd.Close()
	defer os.Remove(fd.Name())

	_, err = io.Copy(fd, os.Stdin)
	if err != nil {
		return err
	}

	_, err = fd.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	stream, err := mio.NewOutStream(fd, ctx.Bool("raw"), key)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, stream)
	return err
}

func handleDebugEncodeStream(ctx *cli.Context) error {
	key, err := readDebugKey(ctx)
	if err != nil {
		return err
	}

	hint := hints.Hint{
		EncryptionAlgo:  hints.EncryptionHint(ctx.String("encryption")),
		CompressionAlgo: hints.CompressionHint(ctx.String("compression")),
	}

	if !hint.IsValid() {
		return fmt.Errorf("invalid encryption or compression")
	}

	r, _, err := mio.NewInStream(os.Stdin, "", key, hint)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, r)
	return err
}

func readStreamSized(ctx *cli.Context) (uint64, error) {
	return humanize.ParseBytes(ctx.String("size"))
}

func handleDebugTenSource(ctx *cli.Context) error {
	s, err := readStreamSized(ctx)
	if err != nil {
		return err
	}

	tr := &testutil.TenReader{}
	_, err = io.Copy(os.Stdout, io.LimitReader(tr, int64(s)))
	return err
}

func handleDebugTenSink(ctx *cli.Context) error {
	s, err := readStreamSized(ctx)
	if err != nil {
		return err
	}

	tw := &testutil.TenWriter{}
	n, err := io.Copy(tw, os.Stdin)
	if err != nil {
		return err
	}

	if int64(s) != n {
		return fmt.Errorf("expected %d, got %d bytes", s, n)
	}

	return nil
}

func handleDebugFuseMock(ctx *cli.Context) error {
	opts := fusetest.Options{
		CatfsPath:           ctx.String("catfs-path"),
		MountPath:           ctx.String("mount-path"),
		IpfsPathOrMultiaddr: ctx.String("ipfs-path-or-multiaddr"),
		URL:                 ctx.String("url"),
		MountReadOnly:       ctx.Bool("mount-ro"),
		MountOffline:        ctx.Bool("mount-offline"),
	}

	return fusetest.Launch(opts)
}
