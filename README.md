# S3 Helper
s3helper is a CLI tool designed for random 1 off commands for use cases that are slightly more complex than what the AWS CLI handles

Installation
---
```
go install github.com/GetTerminus/s3helper
```

Usage
---
```
s3helper -h
```

Compiling
---
1. [Install Go](https://golang.org/doc/install) (On OSX you can run `brew install go`)
2. [Install golang/dep](https://github.com/golang/dep): `go get -u github.com/golang/dep/cmd/dep`
3. Clone this repository to your GOPATH: `go get github.com/GetTerminus/s3helper`
4. Run `dep ensure` from the repo's root dir to install the necessary dependencies
5. Run `go build` to build the source code

Troubleshooting
---
If you encounter problems using `go get` to download a package over ssh, it may be necessary to create a rewrite rule in git to rewrite https requests to use ssh. This can be done with the following command:
```
git config --global url."git@github.com".insteadOf "https://github.com/"
```
