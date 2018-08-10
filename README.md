# Cloud-Key-Client

Cloud-Key-Client is a Golang client that connects up to cloud providers either to collect details of
Service Account keys, or manipulate them.

Amongst the details collected is the age (in minutes) of each key. This could
prove useful for applications that apply further analysis and/or processing on
them, such as key rotation.

The following providers have been integrated:

* AWS
* GCP

No config is required, you simply need to pass a slice of `providerRequest`
structs to the `keys()` func.

Authorisation is handled by the Default Credential Provider Chains for both
[GCP](https://cloud.google.com/docs/authentication/production#auth-cloud-implicit-go) and [AWS](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default).
