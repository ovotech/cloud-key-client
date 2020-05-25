package main

import (
	"fmt"

	"github.com/ovotech/cloud-key-client/pkg/keys"
)

func main() {
	var providers []keys.Provider
	providers = append(providers,
		keys.Provider{
			Provider:   "gcp",
			GcpProject: "my-project",
		})
	keys, err := keys.Keys(providers, false)
	if err == nil {
		for _, key := range keys {
			fmt.Println(key.ID)
		}
	}

}
