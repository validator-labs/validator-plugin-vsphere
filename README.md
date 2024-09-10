[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/validator-labs/validator-plugin-vsphere/issues)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Test](https://github.com/validator-labs/validator-plugin-vsphere/actions/workflows/test.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/validator-labs/validator-plugin-vsphere)](https://goreportcard.com/report/github.com/validator-labs/validator-plugin-vsphere)
[![codecov](https://codecov.io/gh/validator-labs/validator-plugin-vsphere/graph/badge.svg?token=QHR08U8SEQ)](https://codecov.io/gh/validator-labs/validator-plugin-vsphere)
[![Go Reference](https://pkg.go.dev/badge/github.com/validator-labs/validator-plugin-vsphere.svg)](https://pkg.go.dev/github.com/validator-labs/validator-plugin-vsphere)

# validator-plugin-vsphere
The vSphere [validator](https://github.com/validator-labs/validator) plugin ensures that your vSphere environment matches a user-configurable expected state.

## Description
The vSphere validator plugin reconciles `VsphereValidator` custom resources to perform the following validations against your vSphere environment:

1. Compare a user's privileges with respect to a particular entity against an expected privilege set.

   Supported entities:
   - Cluster, Datacenter, Datastore, Folder, ESXi Host, Network, Resource Pool, vApp, vCenter root, Distributed Port Group, Distributed Switch

   Required Privileges:
   - `System.View`
2. Check if sufficient compute resources are available on a particular entity to satify a resource request.

   Supported entities:
   - Cluster, ESXi Host, Resource Pool

   Required Privileges:
   - TODO: identify and update
3. Compare the tags associated with a particular entity against an expected tag set.

   Supported entities:
   - Cluster, Datacenter, ESXi Host, Resource Pool, VM

   Required Privileges:
   - - TODO: identify and update
4. Check if a given set of ESXi Hosts all have NTP enabled and running, with identical NTP servers configured.

   Required Privileges:
   - - TODO: identify and update

Each `VsphereValidator` CR is (re)-processed every two minutes to continuously ensure that your vSphere environment matches the expected state.

See the [samples](https://github.com/validator-labs/validator-plugin-vsphere/tree/main/config/samples) directory for example `VsphereValidator` configurations.

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
All contributions are welcome! Feel free to reach out on the [Spectro Cloud community Slack](https://spectrocloudcommunity.slack.com/join/shared_invite/zt-g8gfzrhf-cKavsGD_myOh30K24pImLA#/shared-invite/email).

Make sure `pre-commit` is [installed](https://pre-commit.com#install).

Install the `pre-commit` scripts:

```console
pre-commit install --hook-type commit-msg
pre-commit install --hook-type pre-commit
```

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

