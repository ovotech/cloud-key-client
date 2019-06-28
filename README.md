# Cloud-Key-Client
 [![CircleCI](https://circleci.com/gh/ovotech/cloud-key-client.svg?style=svg&circle-token=4a7b48b664bf017b6256234f5de24c5b70c54168)](https://circleci.com/gh/ovotech/cloud-key-client)

Cloud-Key-Client is a Golang client that connects up to cloud providers either
to collect details of Service Account keys, or manipulate them.


## Install as a Go Dependency

```go
go get -u github.com/ovotech/cloud-key-client
```


## Purpose

The data of Service Account Keys that the client can return:

```go
//Key type
type Key struct {
	Account       string
	FullAccount   string
	Age           float64 //minutes
	ID            string
	LifeRemaining float64
	Name          string
	Provider      Provider
}

//Provider type
type Provider struct {
	Provider   string //e.g. "aws" or "gcp"
	GcpProject string //Required only when using GCP
}
```

Note the age of each key is returned. This could prove useful for users who want
to track the ages of their keys, alert on old keys, and/or rotate them.


## Integrations

The following providers have been integrated:

* AWS
* GCP

No config is required, you simply need to pass a slice of `providerRequest`
structs to the `keys()` func.

Authorisation is handled by the Default Credential Provider Chains for both
[GCP](https://cloud.google.com/docs/authentication/production#auth-cloud-implicit-go) and [AWS](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default).
