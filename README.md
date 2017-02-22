# cliaas
a cli tool to help with iaas interactions (this is meant to be a single cli for all iaases, instead of pulling in each individual tool for each iaas)


unit test status: [![concourse.customer0.io](https://concourse.customer0.io/api/v1/teams/pcfs/pipelines/cliaas-ci/jobs/unit-tests/badge)](https://concourse.customer0.io/teams/pcfs/pipelines/cliaas-ci) 
integration test status: [![concourse.customer0.io](https://concourse.customer0.io/api/v1/teams/pcfs/pipelines/cliaas-ci/jobs/integration-tests/badge)](https://concourse.customer0.io/teams/pcfs/pipelines/cliaas-ci) 

## CI Pipeline

The `integration-tests` job can read GCP credentials through a params file or can be passed in via the command line as follows to avoid copying data from the credentials JSON:


```
fly -t c0 set-pipeline -p cliaas-ci -c ci/pipeline.yml -l ci/params.yml --var gcp_creds="$(cat /path/to/creds.json)"
```
