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

`cliaas -c config.yml replace-vm --identifier vm-identifier --image some-image`

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
EOF
```

* `access_key_id`: The AWS_ACCESS_KEY_ID to use. Must have the ability to stop VM, start VM, and associate an IP address.
* `secret_access_key`: The AWS_SECRET_ACCESS_KEY to use. Must have the ability to stop VM, start VM, and associate an IP address.
* `region`: The AWS region to use.
* `vpc`: The AWS vpc to use.
* `ami`: A Pivotal Cloud Foundry Operations Manager AMI, for the new VM in `replace-vm`.

#### GCP-specific Config

```
cat > config.yml <<EOF
  gcp:
    credfile: /tmp/gcp-creds.json
    zone: us-east-1
    project: my-gcp-projectname
EOF
```

* `credfile`: The path of your credentials json file issued by gcp.
* `zone`: the zone in gcp your deployments are in.
* `project`: the name of the gcp project you're using.

#### Azure-specific Config

```
cat > config.yml <<EOF
  azure:
    subscription_id: xxxxxx 
    client_id: xxxxx
    client_secret: xxxxx
    tenant_id: xxxxx
    resource_group_name: xxxxx
    resource_manager_endpoint: xxxxx
EOF
```

* `subscription_id`: xxxxxx //$ azure account show | grep "data: ID" 
* `client_id`: xxxxx //this is the appID output of azure ad app create
* `client_secret`: xxxxx //this is the password given as a param to the $ azure ad sp create
* `tenant_id`: xxxxx //$ azure account show | grep "Tenant ID"
* `resource_group_name`: xxxxx /// this is just the resource group name we wish to target
* `resource_manager_endpoint`: xxxx /// this is an option value. defaults to https://management.azure.com/

#### Identifiers

The VM identifier is used to find the VM by name in the IaaS.

#### Images

* For AWS, the image is an AMI, e.g. ami-019e4617
* For GCP, the image is a disk image url, e.g. https://storage.googleapis.com/ops-manager-us/pcf-gcp-1.9.3.tar.gz

## Developing

```
go install github.com/onsi/ginkgo/ginkgo
ginkgo -r -p -race -skipPackage integration
```
