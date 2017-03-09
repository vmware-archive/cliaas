# cliaas

`cliaas` wraps multiple IaaS-specific libraries to perform some IaaS-agnostic
functions. Presently it only supports upgrading a Pivotal Cloud Foundry
Operations Manager VM.

## Installing

Download the [latest release](https://github.com/pivotal-cf/cliaas/releases/latest).

### Install from source

Requirements:

* [glide](https://github.com/masterminds/glide)
* [go](https://golang.org)

```
go get github.com/pivotal-cf/cliaas
cd $GOPATH/src/github.com/pivotal-cf/cliaas
glide install
go install github.com/pivotal-cf/cliaas/cmd/cliaas
```

## Usage

`cliaas -c config.yml replace-vm -i vm-identifier`

### Config

The `-c, --config=` flag is for specifying a YAML file with IaaS-specific configuration options to use when running a command. The config should only contain the configuration for a single IaaS for now.

#### AWS-specific Config

```
cat > config.yml <<EOF
  aws:
    access_key_id: example-access-key-id
    secret_access_key: example-secret-access-key
    region: us-east-1
    vpc: vpc-12345678
    ami: ami-019e4617
EOF
```

* `access_key_id`: The AWS_ACCESS_KEY_ID to use. Must have the ability to stop VM, start VM, and associate an IP address.
* `secret_access_key`: The AWS_SECRET_ACCESS_KEY to use. Must have the ability to stop VM, start VM, and associate an IP address.
* `region`: The AWS region to use.
* `vpc`: The AWS vpc to use.
* `ami`: A Pivotal Cloud Foundry Operations Manager AMI, for the new VM in `replace-vm`.

## Developing

```
go install github.com/onsi/ginkgo/ginkgo
ginkgo -r -p -race -skipPackage integration
```
