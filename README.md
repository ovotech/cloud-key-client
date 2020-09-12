# Cloud-Key-Client
 [![CircleCI](https://circleci.com/gh/ovotech/cloud-key-client.svg?style=svg&circle-token=4a7b48b664bf017b6256234f5de24c5b70c54168)](https://circleci.com/gh/ovotech/cloud-key-client)

Cloud-Key-Client is a Golang client that connects up to cloud providers either
to collect details of Service Account keys, or manipulate them.


## Install as a Go Dependency

```go
go get -u github.com/ovotech/cloud-key-client
```


## Getting Started

```go
package main

import (
	"fmt"

	keys "github.com/ovotech/cloud-key-client"
)

func main() {
	providers := []keys.Provider{}

	// create a GCP provider
	gcpProvider := keys.Provider{
		GcpProject: "pe-dev-185509",
		Provider:   "gcp",
	}
	// create an AWS provider
	awsProvider := keys.Provider{
		Provider: "aws",
	}

	// add both providers to the slice
	providers = append(providers, gcpProvider)
	providers = append(providers, awsProvider)

	// use the cloud-key-client
	keys, err := keys.Keys(providers, true)
	if err != nil {
		fmt.Print(err)
		return
	}
	for _, key := range keys {
		fmt.Printf("%s, ID: ****%s, Age: %dd, Status: %s\n",
			key.Account,
			key.ID[len(key.ID)-4:],
			int(key.Age/1440),
			key.Status)
	}
}
```

## Purpose

This client could be useful for obtaining key metadata, such as age, and 
performing create and delete operations for key rotation. Multiple providers 
can be accessed through a single interface.


## Integrations

The following cloud providers have been integrated:

* AWS
* GCP

No config is required, you simply need to pass a slice of `Provider` structs to
the `keys()` func.

Authentication is handled by the Default Credential Provider Chains for both
[GCP](https://cloud.google.com/docs/authentication/production#auth-cloud-implicit-go)
and [AWS](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default).
