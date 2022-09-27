# application-exporter
Go application to export the configuration of applications deployed in OpenShift.
* Filter namespaces by configurable label(s)
* For each application (e.g., any `Deplopyment`, `DeploymentConfig` and `StatefulSet` in the matching namespaces), collect the image name and version and the resource configuration and usage (optional)
* Export configuration in configurable format (text or CSV)
* Run as a script or a REST service
* Run as a standalone executable or in OpenShift containerized environment (REST service only)

Sample output in CSV format without the resource configuration and usage data:

|namespace | application | container | imageName | imageVersion | fullImageName|
|---|---|---|---|---|---|
|rhpam | rhpam-authoring-rhpamcentr | rhpam-authoring-rhpamcentr | rhpam-businesscentral-rhel8 | 7.9.1 | image-registry.openshift-image-registry.svc:5000/rhpam/rhpam-businesscentral-rhel8@sha256:38172680f719cd8eeff1fdf4f2732e7cfdea5109d381ef9108e1c88b74390bc5|
|rhpam | rhpam-server | rhpam-server | rhpam-server | 7.9.1 | image-registry.openshift-image-registry.svc:5000/rhpam/rhpam-server@sha256:7f2df7e673e1e9def8575026ef4697341227a9d5860bcb6d3101d80a0701dd3e|

Sample output in CSV format including the resource configuration and usage data:
|namespace | application |  container |  imageName |  imageVersion |  fullImageName |  CPU limits |  memory limits |  CPU requests |  memory requests |  pod |  CPU usage |  memory usage|
|---|---|---|---|---|---|---|---|---|---|---|---|---|
|rhpam | rhpam-authoring-rhpamcentr | rhpam-authoring-rhpamcentr | rhpam-businesscentral-rhel8 | 7.9.1 | image-registry.openshift-image-registry.svc:5000/rhpam/rhpam-businesscentral-rhel8@sha256:38172680f719cd8eeff1fdf4f2732e7cfdea5109d381ef9108e1c88b74390bc5 | 2 | 4Gi | 1500m | 3Gi | rhpam-authoring-rhpamcentr-1-jqq2l | 5m | 1493208Ki|
|rhpam | rhpam-server | rhpam-server | rhpam-server | 7.9.1 | image-registry.openshift-image-registry.svc:5000/rhpam/rhpam-server@sha256:7f2df7e673e1e9def8575026ef4697341227a9d5860bcb6d3101d80a0701dd3e | 1 | 2Gi | 750m | 1536Mi | rhpam-server-22-4lhwt | 2m | 1058236Ki|

## CI pipeline
A GitHub action runs at every new release, and generates the following artifacts:
* The `inventory-exporter.tar` artifact is added to the [release page](https://github.com/dmartinol/application-exporter/releases) after some time
* The updated container image is published on the public repository [quay.io/dmartino](https://quay.io/dmartino)

The version is printed at the application startup, as in:
```bash
2022-09-23T17:21:58.411+0200	info	The version of ./bin/inventory-exporter-darwin-amd64 is : 0.1.1
```

## Configurable options
### Command line arguments
```bash
Usage of ./application-exporter:
  -as-service
        Run as REST service
  -content-type string
        Content type, one of text, CSV (default "text")
  -log-level string
        Log level, one of debug, info, warn (default "info")
  -ns-selector string
        Namespace selector, like label1=value1,label2=value2
  -output string
        Output file name, default is output.<content-type>. File suffix is automatically added
  -server-port int
        Server port (only for REST service mode) (default 8080)
  -with-resources
        Include resource configuration and usage
```

### Environment variables
The following environment variables can override the command arguments:
* `AS_SERVICE`: any value, overrides `-as-service` command line argument
* `IN_CONTAINER`: any value
* `LOG_LEVEL`: overrides `-log-level` command line argument
* `NS_SELECTOR`: overrides `-ns-selector` command line argument
* `CONTENT_TYPE`: overrides `-content-type` command line argument
* `SERVER_PORT`: overrides `-server-port` command line argument

### REST query 
The following query parameters can override the command arguments and environment variables:
* `content-type`: overrides `-content-type` command line argument and `CONTENT_TYPE` environment variable
* `ns-selector`: overrides `-ns-selector` command line argument and `NS_SELECTOR` environment variable
* `output`: overrides `-output` command line argument
* `with-resources`: any value, overrides `-with-resources` command line argument

## Running as standalone executable
### Running with `go run`
Requirements:
* Active OpenShift login
* `go`, at least version `1.19`
* Clone this git repository

Run the following from the root folder of the cloned repository:
```
go run main.go
```
Show the available options with:
```
go run main.go --help
```

The following are examples of requests performed using `curl`:
```bash
curl "http://localhost:8080/inventory"
curl "http://localhost:8080/inventory?content-type=CSV&ns-selector=mylabel=myvalue"
```

### Running the binary executable
Requirements:
* Active OpenShift login
* Download the latest binaries from the [release page](https://github.com/dmartinol/application-exporter/releases) and extract the content

Run the following from the folder where you extracted the release archive:
```
./bin/inventory-exporter-darwin-amd64
```
Note: the file name might change depending on the actual target machine and OC.

All the [command line arguments](#command-line-arguments) and [environment variables](#environment-variables) described before are also applicable.

#### Building the executable
According to your target platform, choose one of the following build commands to manually build the executable for a given custom version:
```bash
export BUILD_VERSION=<YOUR_VERSION>
GOOS=windows GOARCH=386 go build -o inventory-exporter-win-386.exe -ldflags "-X main.BuildVersion=$BUILD_VERSION" main.go
GOOS=windows GOARCH=amd64 go build -o inventory-exporter-win-amd64.exe -ldflags "-X main.BuildVersion=$BUILD_VERSION" main.go
GOOS=darwin GOARCH=amd64 go build -o inventory-exporter-darwin-amd64 -ldflags "-X main.BuildVersion=$BUILD_VERSION" main.go
GOOS=darwin GOARCH=arm64 go build -o inventory-exporter-darwin-arm64 -ldflags "-X main.BuildVersion=$BUILD_VERSION" main.go
```

## Running in OpenShift
Requirements:
* Active OpenShift login

OpenShift [templates](./openshift/) are available to simplify the deployment in the containerized environment.

### Running as a regular Service
* Using an existing image
```bash
export APP_NAMESPACE=exporter
export APP_IMAGE=quay.io/dmartino/application-exporter:0.1.1
oc project ${APP_NAMESPACE}
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/rbac.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=${APP_IMAGE} -f openshift/service.yaml | oc apply -f -
```
* Building the image in the local registry from the Git repo:
```bash
export APP_NAMESPACE=exporter
export APP_IMAGE=image-registry.openshift-image-registry.svc:5000/exporter/application-exporter:latest
oc project ${APP_NAMESPACE}
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/rbac.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/build.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=${APP_IMAGE} -f openshift/service.yaml | oc apply -f -
```

Run `oc get route -n ${APP_NAMESPACE} inventory-exporter` to get the public URL of your service. Add the `/inventory` path 
before invoking the services.

### Running as a Serverless Service
Requirements:
* `Red Hat Serverless` operator installed and configured
  * At least the `KnativeServing` instance is needed

* Using an existing image
```bash
export APP_NAMESPACE=exporter
export APP_IMAGE=quay.io/dmartino/application-exporter:0.1.0
oc project ${APP_NAMESPACE}
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/rbac.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=${APP_IMAGE} -f openshift/serverless.yaml | oc apply -f -
```
* Building the image in the local registry from the Git repo:
```bash
export APP_NAMESPACE=exporter
export APP_IMAGE=image-registry.openshift-image-registry.svc:5000/exporter/application-exporter:latest
oc project ${APP_NAMESPACE}
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/rbac.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/build.yaml | oc apply -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=${APP_IMAGE} -f openshift/serverless.yaml | oc apply -f -
```

Run `oc get ksvc -n ${APP_NAMESPACE} application-exporter-knative` to get the public URL of your service. Add the `/inventory` path 
before invoking the services.

### Optional template parameters
The following parameters in [build.yaml](./openshift/build.yaml) have default values and don't usually need to be customized:
```yaml
- description: Git repo
  from: '[A-Z0-9]{8}'
  generate: expression
  name: GIT_REPO
  value: https://github.com/dmartinol/application-exporter.git
- description: Git ref
  from: '[A-Z0-9]{8}'
  generate: expression
  name: GIT_REF
  value: main
- description: Build version
  from: '[A-Z0-9]{8}'
  generate: expression
  name: BUILD_VERSION
  value: latest
```

### Uninstalling the application
Run the following commands to completely uninstall the application:
```
export APP_NAMESPACE=exporter
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/rbac.yaml | oc delete -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -f openshift/build.yaml | oc delete -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=NA -f openshift/service.yaml | oc delete -f -
oc process -p=APP_NAMESPACE=${APP_NAMESPACE} -p=APP_IMAGE=NA -f openshift/serverless.yaml | oc delete -f -
```

## Open issues
See [here](https://github.com/dmartinol/application-exporter/issues)

## License
The source code and documentation in this project are released under the [Apache 2.0 license](./LICENSE).