// package shell implements a remote API interface for a running ipfs daemon
package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	"os"
	"path"
	"strings"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	tar "github.com/whyrusleeping/tar-utils"
)

const (
	DefaultPathName = ".ipfs"
	DefaultPathRoot = "~/" + DefaultPathName
	DefaultApiFile  = "api"
	EnvDir          = "IPFS_PATH"
)

type Shell struct {
	url     string
	httpcli gohttp.Client
}

func NewLocalShell() *Shell {
	baseDir := os.Getenv(EnvDir)
	if baseDir == "" {
		baseDir = DefaultPathRoot
	}

	baseDir, err := homedir.Expand(baseDir)
	if err != nil {
		return nil
	}

	apiFile := path.Join(baseDir, DefaultApiFile)

	if _, err := os.Stat(apiFile); err != nil {
		return nil
	}

	api, err := ioutil.ReadFile(apiFile)
	if err != nil {
		return nil
	}

	return NewShell(strings.TrimSpace(string(api)))
}

func NewShell(url string) *Shell {
	c := &gohttp.Client{
		Transport: &gohttp.Transport{
			Proxy:             gohttp.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	}

	return NewShellWithClient(url, c)
}

func NewShellWithClient(url string, c *gohttp.Client) *Shell {
	if a, err := ma.NewMultiaddr(url); err == nil {
		_, host, err := manet.DialArgs(a)
		if err == nil {
			url = host
		}
	}
	var sh Shell
	sh.url = url
	sh.httpcli = *c
	// We don't support redirects.
	sh.httpcli.CheckRedirect = func(_ *gohttp.Request, _ []*gohttp.Request) error {
		return fmt.Errorf("unexpected redirect")
	}
	return &sh
}

func (s *Shell) SetTimeout(d time.Duration) {
	s.httpcli.Timeout = d
}

func (s *Shell) Request(command string, args ...string) *RequestBuilder {
	return &RequestBuilder{
		command: command,
		args:    args,
		shell:   s,
	}
}

type IdOutput struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
}

// ID gets information about a given peer.  Arguments:
//
// peer: peer.ID of the node to look up.  If no peer is specified,
//   return information about the local peer.
func (s *Shell) ID(peer ...string) (*IdOutput, error) {
	if len(peer) > 1 {
		return nil, fmt.Errorf("Too many peer arguments")
	}

	var out IdOutput
	if err := s.Request("id", peer...).Exec(context.Background(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cat the content at the given path. Callers need to drain and close the returned reader after usage.
func (s *Shell) Cat(path string) (io.ReadCloser, error) {
	resp, err := s.Request("cat", path).Send(context.Background())
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

const (
	TRaw = iota
	TDirectory
	TFile
	TMetadata
	TSymlink
)

// List entries at the given path
func (s *Shell) List(path string) ([]*LsLink, error) {
	var out struct{ Objects []LsObject }
	err := s.Request("ls", path).Exec(context.Background(), &out)
	if err != nil {
		return nil, err
	}
	if len(out.Objects) != 1 {
		return nil, errors.New("bad response from server")
	}
	return out.Objects[0].Links, nil
}

type LsLink struct {
	Hash string
	Name string
	Size uint64
	Type int
}

type LsObject struct {
	Links []*LsLink
	LsLink
}

// Pin the given path
func (s *Shell) Pin(path string) error {
	return s.Request("pin/add", path).
		Option("recursive", true).
		Exec(context.Background(), nil)
}

// Unpin the given path
func (s *Shell) Unpin(path string) error {
	return s.Request("pin/rm", path).
		Option("recursive", true).
		Exec(context.Background(), nil)
}

const (
	DirectPin    = "direct"
	RecursivePin = "recursive"
	IndirectPin  = "indirect"
)

type PinInfo struct {
	Type string
}

// Pins returns a map of the pin hashes to their info (currently just the
// pin type, one of DirectPin, RecursivePin, or IndirectPin. A map is returned
// instead of a slice because it is easier to do existence lookup by map key
// than unordered array searching. The map is likely to be more useful to a
// client than a flat list.
func (s *Shell) Pins() (map[string]PinInfo, error) {
	var raw struct{ Keys map[string]PinInfo }
	return raw.Keys, s.Request("pin/ls").Exec(context.Background(), &raw)
}

type PeerInfo struct {
	Addrs []string
	ID    string
}

func (s *Shell) FindPeer(peer string) (*PeerInfo, error) {
	var peers struct{ Responses []PeerInfo }
	err := s.Request("dht/findpeer", peer).Exec(context.Background(), &peers)
	if err != nil {
		return nil, err
	}
	if len(peers.Responses) == 0 {
		return nil, errors.New("peer not found")
	}
	return &peers.Responses[0], nil
}

func (s *Shell) Refs(hash string, recursive bool) (<-chan string, error) {
	resp, err := s.Request("refs", hash).
		Option("recursive", recursive).
		Send(context.Background())
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		resp.Close()
		return nil, resp.Error
	}

	out := make(chan string)
	go func() {
		defer resp.Close()
		var ref struct {
			Ref string
		}
		defer close(out)
		dec := json.NewDecoder(resp.Output)
		for {
			err := dec.Decode(&ref)
			if err != nil {
				return
			}
			if len(ref.Ref) > 0 {
				out <- ref.Ref
			}
		}
	}()

	return out, nil
}

func (s *Shell) Patch(root, action string, args ...string) (string, error) {
	var out object
	return out.Hash, s.Request("object/patch/"+action, root).
		Arguments(args...).
		Exec(context.Background(), &out)
}

func (s *Shell) PatchData(root string, set bool, data interface{}) (string, error) {
	var read io.Reader
	switch d := data.(type) {
	case io.Reader:
		read = d
	case []byte:
		read = bytes.NewReader(d)
	case string:
		read = strings.NewReader(d)
	default:
		return "", fmt.Errorf("unrecognized type: %#v", data)
	}

	cmd := "append-data"
	if set {
		cmd = "set-data"
	}

	fr := files.NewReaderFile(read)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	var out object
	return out.Hash, s.Request("object/patch/"+cmd, root).
		Body(fileReader).
		Exec(context.Background(), &out)
}

func (s *Shell) PatchLink(root, path, childhash string, create bool) (string, error) {
	var out object
	return out.Hash, s.Request("object/patch/add-link", root, path, childhash).
		Option("create", create).
		Exec(context.Background(), &out)
}

func (s *Shell) Get(hash, outdir string) error {
	resp, err := s.Request("get", hash).Option("create", true).Send(context.Background())
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}

	extractor := &tar.Extractor{Path: outdir}
	return extractor.Extract(resp.Output)
}

func (s *Shell) NewObject(template string) (string, error) {
	var out object
	req := s.Request("object/new")
	if template != "" {
		req.Arguments(template)
	}
	return out.Hash, req.Exec(context.Background(), &out)
}

func (s *Shell) ResolvePath(path string) (string, error) {
	var out struct {
		Path string
	}
	err := s.Request("resolve", path).Exec(context.Background(), &out)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(out.Path, "/ipfs/"), nil
}

// returns ipfs version and commit sha
func (s *Shell) Version() (string, string, error) {
	ver := struct {
		Version string
		Commit  string
	}{}

	if err := s.Request("version").Exec(context.Background(), &ver); err != nil {
		return "", "", err
	}
	return ver.Version, ver.Commit, nil
}

func (s *Shell) IsUp() bool {
	_, _, err := s.Version()
	return err == nil
}

func (s *Shell) BlockStat(path string) (string, int, error) {
	var inf struct {
		Key  string
		Size int
	}

	if err := s.Request("block/stat", path).Exec(context.Background(), &inf); err != nil {
		return "", 0, err
	}
	return inf.Key, inf.Size, nil
}

func (s *Shell) BlockGet(path string) ([]byte, error) {
	resp, err := s.Request("block/get", path).Send(context.Background())
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	return ioutil.ReadAll(resp.Output)
}

func (s *Shell) BlockPut(block []byte, format, mhtype string, mhlen int) (string, error) {
	var out struct {
		Key string
	}

	fr := files.NewBytesFile(block)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	return out.Key, s.Request("block/put").
		Option("mhtype", mhtype).
		Option("format", format).
		Option("mhlen", mhlen).
		Body(fileReader).
		Exec(context.Background(), &out)
}

type IpfsObject struct {
	Links []ObjectLink
	Data  string
}

type ObjectLink struct {
	Name, Hash string
	Size       uint64
}

func (s *Shell) ObjectGet(path string) (*IpfsObject, error) {
	var obj IpfsObject
	if err := s.Request("object/get", path).Exec(context.Background(), &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (s *Shell) ObjectPut(obj *IpfsObject) (string, error) {
	var data bytes.Buffer
	err := json.NewEncoder(&data).Encode(obj)
	if err != nil {
		return "", err
	}

	fr := files.NewReaderFile(&data)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	var out object
	return out.Hash, s.Request("object/put").
		Body(fileReader).
		Exec(context.Background(), &out)
}

func (s *Shell) PubSubSubscribe(topic string) (*PubSubSubscription, error) {
	// connect
	resp, err := s.Request("pubsub/sub", topic).Send(context.Background())
	if err != nil {
		return nil, err
	}
	return newPubSubSubscription(resp), nil
}

func (s *Shell) PubSubPublish(topic, data string) (err error) {
	resp, err := s.Request("pubsub/pub", topic, data).Send(context.Background())
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

type ObjectStats struct {
	Hash           string
	BlockSize      int
	CumulativeSize int
	DataSize       int
	LinksSize      int
	NumLinks       int
}

// ObjectStat gets stats for the DAG object named by key. It returns
// the stats of the requested Object or an error.
func (s *Shell) ObjectStat(key string) (*ObjectStats, error) {
	var stat ObjectStats
	err := s.Request("object/stat", key).Exec(context.Background(), &stat)
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

type Stats struct {
	TotalIn  int64
	TotalOut int64
	RateIn   float64
	RateOut  float64
}

// ObjectStat gets stats for the DAG object named by key. It returns
// the stats of the requested Object or an error.
func (s *Shell) StatsBW(ctx context.Context) (*Stats, error) {
	v := &Stats{}
	err := s.Request("stats/bw").Exec(ctx, &v)
	return v, err
}

type SwarmStreamInfo struct {
	Protocol string
}

type SwarmConnInfo struct {
	Addr    string
	Peer    string
	Latency string
	Muxer   string
	Streams []SwarmStreamInfo
}

type SwarmConnInfos struct {
	Peers []SwarmConnInfo
}

// SwarmPeers gets all the swarm peers
func (s *Shell) SwarmPeers(ctx context.Context) (*SwarmConnInfos, error) {
	v := &SwarmConnInfos{}
	err := s.Request("swarm/peers").Exec(ctx, &v)
	return v, err
}

type swarmConnection struct {
	Strings []string
}

// SwarmConnect opens a swarm connection to a specific address.
func (s *Shell) SwarmConnect(ctx context.Context, addr ...string) error {
	var conn *swarmConnection
	err := s.Request("swarm/connect").
		Arguments(addr...).
		Exec(ctx, &conn)
	return err
}
