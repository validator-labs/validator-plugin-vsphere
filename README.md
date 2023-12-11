[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Build](https://github.com/spectrocloud-labs/validator-plugin-vsphere/actions/workflows/build_container.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/spectrocloud-labs/validator-plugin-vsphere)](https://goreportcard.com/report/github.com/spectrocloud-labs/validator-plugin-vsphere)
[![codecov](https://codecov.io/gh/spectrocloud-labs/validator-plugin-vsphere/graph/badge.svg?token=QHR08U8SEQ)](https://codecov.io/gh/spectrocloud-labs/validator-plugin-vsphere)
[![Go Reference](https://pkg.go.dev/badge/github.com/spectrocloud-labs/validator-plugin-vsphere.svg)](https://pkg.go.dev/github.com/spectrocloud-labs/validator-plugin-vsphere)

# validator-plugin-vsphere
The vSphere [validator](https://github.com/spectrocloud-labs/validator) plugin ensures that your vSphere environment matches a user-configurable expected state.

## Description
The vSphere validator plugin reconciles `VsphereValidator` custom resources to perform the following validations against your vSphere environment:

1. Compare the privileges associated with a user against an expected privileges set
2. Compare the privileges associated with a user against an expected privileges set on a particular entity(cluster, resourcepool, folder, vapp, host)
3. Check if enough compute resources are available on a host, resourcepool or cluster against a resource request
4. Compare the tags associated with a datacenter, cluster, host, vm, resourcepool or vm against an expected tag set
5. Check if a given set of host systems have a valid NTP configuration

Each `VsphereValidator` CR is (re)-processed every two minutes to continuously ensure that your vSphere environment matches the expected state.

See the [samples](https://github.com/spectrocloud-labs/validator-plugin-vsphere/tree/main/config/samples) directory for example `VsphereValidator` configurations.

> [!NOTE]
> This plugin currently require a user with administrator role to perform all of the validations specified above. Further information on fine-grained permissions required by each validation will be updated in the future.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/validator-plugin-vsphere:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/validator-plugin-vsphere:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

