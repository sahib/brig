package repo

import (
	"fmt"

	yml "gopkg.in/yaml.v2"
)

func buildMetaDefault(backendName, owner string) []byte {
	data, err := yml.Marshal(map[string]interface{}{
		"data": map[string]string{
			"backend": backendName,
		},
		"repo": map[string]string{
			"owner": owner,
		},
	})

	if err != nil {
		panic(fmt.Sprintf("Failed to convert default meta to yml: %v", err))
	}

	return data
}
