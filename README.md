# cliaas
a cli tool to help with iaas interactions (this is meant to be a single cli for
all iaases, instead of pulling in each individual tool for each iaas)


unit test status:
[![concourse.customer0.io](https://concourse.customer0.io/api/v1/teams/pcfs/pipelines/cliaas-ci/jobs/unit-tests/badge)](https://concourse.customer0.io/teams/pcfs/pipelines/cliaas-ci)

integration test status:
[![concourse.customer0.io](https://concourse.customer0.io/api/v1/teams/pcfs/pipelines/cliaas-ci/jobs/integration-tests/badge)](https://concourse.customer0.io/teams/pcfs/pipelines/cliaas-ci)

## Configuration

The cliaas tool takes configuration in the form of environment variables.
The following must be set depending on which IaaS provider you wish to target:

### AWS

  - `AWS_ACCESS_KEY`
  - `AWS_SECRET_KEY`
  - `AWS_REGION`
  - `AWS_VPC`

#### CLI ####
```
aws command options]
          --accesskey=    aws access key [$AWS_ACCESSKEY]
          --secretkey=    aws secret access key [$AWS_SECRETKEY]
          --region=       aws region (default: us-east-1) [$AWS_REGION]
          --vpc=          aws VPC id [$AWS_VPC]
          --name=         aws name tag for vm [$AWS_NAME]
          --ami=          aws ami to provision [$AWS_AMI]
          --instanceType= aws instance type to provision [$AWS_INSTANCE_TYPE]
          --elastic-ip=   aws elastic ip to associate to provisioned VM [$AWS_ELASTIC_IP]
```

### GCP

 - `GCP_CREDS` - String representation of the creds file
 - `GCP_PROJECT`
 - `GCP_ZONE`

## CI Pipeline

The `integration-tests` job can read GCP credentials through a params file or
can be passed in via the command line as follows to avoid copying data from the
credentials JSON:


```
fly -t c0 set-pipeline -p cliaas-ci -c ci/pipeline.yml -l ci/params.yml --var
gcp_creds="$(cat /path/to/creds.json)"
```

### Integration Tests

The `IAAS` environment variable can be used to target a specific providers
integration tests when running `integration_tests/task.yml`.
