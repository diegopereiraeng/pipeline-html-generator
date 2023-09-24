# Pipeline HTML Generator

## Description

This is a Go-based microservice that generates HTML reports for your CI/CD pipeline executions. It is designed to work seamlessly with Harness.io's CD pipelines.

## Features

- Generate HTML reports for pipeline status
- Customizable report parameters
- Integration with Harness.io

## Prerequisites

- Go version 1.x
- Docker
- Harness account
- Kubernetes cluster (for KubernetesDirect stepGroupInfra)

## Installation

Clone the repository:

```bash
git clone https://github.com/diegopereiraeng/pipeline-html-generator.git
```

Build the Go application:

```bash
go build -o pipeline-html-generator
```

## Usage

Run the application locally:

```bash
./pipeline-html-generator --acc_id=<account_id> --org_id=<org_id> --project_id=<project_id> --pipeline_id=<pipeline_id> --status_list=<status_list> --repo_name=<repo_name> --branch=<branch> --harness_secret=<harness_secret>
```

### Example

```bash
./pipeline-html-generator --acc_id=2_Trfzo9Qeu9fXvj-AcbCQ --org_id=default --project_id=GIT_FLOW_DEMO --pipeline_id=Banking_Validation_Pipeline --status_list=Success --repo_name=payments-validation --branch=master  --harness_secret=pat.2_Trfzo9Qeu9fXvj-gtyXd.76drg4Yhpcd3615245670s6h.D5zxCoRgt5UgE7HJ3saE
```

## Harness CI Integration

``` yaml
- step:
    type: Plugin
    name: Pipeline HTML Generator
    identifier: Pipeline_HTML_Generator
    spec:
    connectorRef: account.DockerHubDiego
    image: diegokoala/pipeline-html-generator:latest
    settings:
        harness_secret: <+secrets.getValue("harnesssatokenplugin")>
    failureStrategies: {}
```

## Harness CD Integration

The service can be integrated into a Harness CD or Custom stage as a plugin (Need Flag: CDS_CONTAINER_STEP_GROUP enabled). Here's a YAML example using a CI step group:

```yaml
- stepGroup:
    name: Dashboard notification
    identifier: Dashboard_notification
    steps:
      - step:
          type: Plugin
          name: Pipeline HTML Generator
          identifier: Pipeline_HTML_Generator
          spec:
            connectorRef: account.DockerHubDiego
            image: diegokoala/pipeline-html-generator:latest
            settings:
              harness_secret: <+secrets.getValue("harnesssatokenplugin")>
              org_id: <+org.identifier>
              project_id: <+project.identifier>
              pipeline_id: <+pipeline.identifier>
              repo_name: <+pipeline.variables.repo_name>
              branch: <+pipeline.properties.ci.codebase.build.spec.branch>
              harness_pipe_execution_url: <+pipeline.executionUrl>
          failureStrategies: {}
    stepGroupInfra:
      type: KubernetesDirect
      spec:
        connectorRef: account.harness_demo
        namespace: harness-delegate-ng
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
