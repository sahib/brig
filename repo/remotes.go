package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	yml "gopkg.in/yaml.v2"
)

type Perms uint32

const (
	PermNone = 0
	PermRead = 1 << iota
	PermWrite
)

type RemotePerms int

func (rp RemotePerms) FromStrings(perms []string) RemotePerms {
	mask := RemotePerms(0)

	for _, perm := range perms {
		switch perm {
		case "read":
			mask |= PermRead
		case "write":
			mask |= PermWrite
		}
	}

	return mask
}

func (rp RemotePerms) ToStrings() []string {
	res := []string{}
	if rp&PermRead > 0 {
		res = append(res, "read")
	}

	if rp&PermWrite > 0 {
		res = append(res, "write")
	}

	return res
}

type Folder struct {
	Folder string
	Perm   RemotePerms
}

func (f Folder) Less(o Folder) bool {
	return f.Folder < o.Folder
}

func insertSortedFolder(folders []Folder, f Folder) []Folder {
	l := len(folders)
	if l == 0 {
		return []Folder{f}
	}

	i := sort.Search(l, func(i int) bool {
		return folders[i].Less(f)
	})

	if i == l {
		return append([]Folder{f}, folders...)
	}

	if i == l-1 {
		return append(folders[0:l], f)
	}

	return append(folders[0:l], append([]Folder{f}, folders[l+1:]...)...)
}

type Remote struct {
	Name    string
	Folders []Folder
}

// RemoteList is a helper that parses the remote access yml file
// and makes it easily accessible from the Go side.
type RemoteList struct {
	remotes map[string]*Remote
}

func NewRemotes(r io.Reader) (*RemoteList, error) {
	remotes := make(map[string]*Remote)
	ymlRemotes := make(map[string][]string)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := yml.Unmarshal(data, ymlRemotes); err != nil {
		return nil, err
	}

	// Go over all remotes denoted in the .yml file.
	for nameAndFolder, perms := range ymlRemotes {
		splitName := strings.Split(nameAndFolder, " ")

		folder := "/"
		name := splitName[0]
		if len(splitName) > 1 {
			folder = splitName[1]
		}

		perms := RemotePerms(0).FromStrings(perms)

		// Append to existing or create new remote.
		if remote, ok := remotes[name]; ok {
			remote.Folders = insertSortedFolder(remote.Folders, Folder{
				Folder: folder,
				Perm:   perms,
			})
		} else {
			remotes[name] = &Remote{
				Name: name,
				Folders: []Folder{{
					Folder: folder,
					Perm:   perms,
				}},
			}
		}
	}

	return &RemoteList{remotes: remotes}, nil
}

func (rl *RemoteList) Export(w io.Writer) error {
	ymlRemotes := make(map[string][]string)

	for _, remote := range rl.remotes {
		for _, folder := range remote.Folders {
			nameAndFolder := strings.Join(
				[]string{remote.Name, folder.Folder},
				" ",
			)

			ymlRemotes[nameAndFolder] = folder.Perm.ToStrings()
		}
	}

	data, err := yml.Marshal(ymlRemotes)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func (rl *RemoteList) AddRemote(remote Remote) error {
	rl.remotes[remote.Name] = &remote
	return nil
}

func (rl *RemoteList) RmRemote(remote Remote) error {
	delete(rl.remotes, remote.Name)
	return nil
}

func (rl *RemoteList) Remote(name string) (Remote, error) {
	rm, ok := rl.remotes[name]
	if !ok {
		return Remote{}, fmt.Errorf("No such remote: %v", name)
	}

	return *rm, nil
}
