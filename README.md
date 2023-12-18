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

### Examples

For examples, see the following subdirectories:

 - [hello-world](examples/tests/hello-world/): a basic example with one slurm cluster to submit jobs to
 - [federated](examples/federated/): more than one cluster connected to the same database.

Note that we don't have pretty rendered docs yet, as this was mostly a quick, few day project, and we are just returning to it to try out federated slurm. If we use or develop beyond a few simple times we will definitely spruce up the docs here.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.


### TODO

- Generate slurm.conf and slurmdbd.conf as templates, with custom hosts, etc.
- Custom user generation?
- If username/password not provided, generate as random
- Add script logging levels / quiet
- consider putting node start in loop (won't exit for job, maybe OK for now)
- make more params in slurm configs variables
- allow the command given to script to be given to srun (timing will be tough, probably need to ensure sinfo working)

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
