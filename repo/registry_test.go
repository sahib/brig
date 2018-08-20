package repo

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	TestRegistryPath = "/tmp/test-registry.yml"
)

func init() {
	RegistryPaths = append(RegistryPaths, TestRegistryPath)
}

func touchTestRegistry(t *testing.T, data []byte) {
	registryFd, err := os.OpenFile(TestRegistryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	require.Nil(t, err)
	defer func() {
		require.Nil(t, registryFd.Close())
	}()

	if data != nil {
		_, err := registryFd.Write(data)
		require.Nil(t, err)
	}

}

func cleanUpTestRegistry(t *testing.T) {
	require.Nil(t, os.Remove(TestRegistryPath))
}

func TestRegistryOpen(t *testing.T) {
	defer cleanUpTestRegistry(t)

	touchTestRegistry(t, nil)
	_, err := OpenRegistry()
	require.Nil(t, err)
}

func TestRegistryAddGet(t *testing.T) {
	defer cleanUpTestRegistry(t)

	touchTestRegistry(t, nil)
	reg, err := OpenRegistry()
	require.Nil(t, err)

	uuid, err := reg.Add(&RegistryEntry{
		Owner:    "owner",
		Password: "password",
		Path:     "/tmp/xxx",
	})

	require.Nil(t, err)
	require.NotEqual(t, uuid, "")

	entry, err := reg.Entry(uuid)
	require.Nil(t, err)
	require.Equal(t, "owner", entry.Owner)
	require.Equal(t, "password", entry.Password)
	require.Equal(t, "/tmp/xxx", entry.Path)
}

func TestRegistryGetEmpty(t *testing.T) {
	defer cleanUpTestRegistry(t)

	touchTestRegistry(t, nil)
	reg, err := OpenRegistry()
	require.Nil(t, err)

	_, err = reg.Entry("")
	require.NotNil(t, err)

	_, err = reg.Entry("nope")
	require.NotNil(t, err)
}

func TestRegistryAddMany(t *testing.T) {
	defer cleanUpTestRegistry(t)

	touchTestRegistry(t, nil)
	reg, err := OpenRegistry()
	require.Nil(t, err)

	for idx := 0; idx < 100; idx++ {
		_, err := reg.Add(&RegistryEntry{
			Owner:    fmt.Sprintf("owner-%d", idx),
			Password: fmt.Sprintf("password-%d", idx),
			Path:     fmt.Sprintf("/tmp/xxx-%d", idx),
		})
		require.Nil(t, err)
	}
}

func TestRegistryOpenTwice(t *testing.T) {
	defer cleanUpTestRegistry(t)

	uuid := ""
	touchTestRegistry(t, nil)

	for i := 0; i < 2; i++ {
		reg, err := OpenRegistry()
		require.Nil(t, err)

		// Only add something on the first run.
		if i == 0 {
			uuid, err = reg.Add(&RegistryEntry{
				Owner:    "owner",
				Password: "password",
				Path:     "/tmp/xxx",
			})

			require.Nil(t, err)
			require.NotEqual(t, uuid, "")
		}

		// Both runs should be able to see the entry.
		entry, err := reg.Entry(uuid)
		require.Nil(t, err)
		require.Equal(t, "owner", entry.Owner)
		require.Equal(t, "password", entry.Password)
		require.Equal(t, "/tmp/xxx", entry.Path)
	}
}
