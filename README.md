# slurm-operator

> What happens when I run out of things to do on a Monday... ohno 

This will be an attempt at creating a slurm operator. I mostly want to learn a production setup for SLURM,
and have some fun! Note that it's not working yet! The next step is to customize the configuration files
(e.g., slurm.conf and slurmdbd.conf) to be config maps, and specific to the cluster.

## Development

### Creation

```bash
mkdir slurm-operator
cd slurm-operator/
operator-sdk init --domain flux-framework.org --repo github.com/converged-computing/slurm-operator
operator-sdk create api --version v1alpha1 --kind slurm --resource --controller
```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

Create a cluster with kind:

```bash
$ kind create cluster
```

You'll need to install the jobset API, which eventually will be added to Kubernetes proper (but is not yet!)

```bash
VERSION=v0.1.3
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```
or development version (this is what I did):

```bash
$ kubectl apply --server-side -k github.com/kubernetes-sigs/jobset/config/default?ref=main
```

Generate the custom resource definition

```bash
# Build and push the image, and generate the examples/dist/slurm-operator-dev.yaml
$ make test-deploy DEVIMG=<some-registry>/slurm-operator:tag

# As an example
$ make test-deploy DEVIMG=vanessa/slurm-operator:test
```

Make our namespace:

```bash
$ kubectl create namespace slurm-operator
```

Apply the new config!

```bash
$ kubectl apply -f examples/dist/slurm-operator-dev.yaml
```

See logs for the operator

```bash
$ kubectl logs -n slurm-operator-system slurm-operator-controller-manager-6f6945579-9pknp 
```

Create a "hello-world" interactive cluster:

```bash
$ kubectl apply -f examples/tests/hello-world/slurm.yaml 
```

When you are done, cleanup.

```bash
$ kind delete cluster
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.


### TODO

- Generate slurm.conf and slurmdbd.conf as templates, with custom hosts, etc.
- Custom user generation?
- If username/password not provided, generate as random
- Add script logging levels / quiet

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
