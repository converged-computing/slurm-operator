# Hello World SLURM

Create a cluster with kind:

```bash
$ kind create cluster
```

You'll need to install the jobset API, which eventually will be added to Kubernetes proper (but is not yet!)

```bash
VERSION=v0.2.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```

or development version (this is what I did) - be careful, this would be release v0.3.0:

```bash
kubectl apply --server-side -k github.com/kubernetes-sigs/jobset/config/default?ref=main
```

Install the SLURM operator:

```bash
# Build and push the image, and generate the examples/dist/slurm-operator-dev.yaml
make test-deploy-recreate DEVIMG=<some-registry>/slurm-operator:tag

# As an example
make test-deploy-recreate DEVIMG=vanessa/slurm-operator:test
```

See logs for the operator

```bash
kubectl logs -n slurm-operator-system slurm-operator-controller-manager-6f6945579-9pknp
```

Wait until you see the operator running. Create a "hello-world" interactive cluster:

```bash
$ kubectl apply -f ./slurm.yaml 
```

Wait until all of the containers are running:

```bash
kubectl get pods
NAME                        READY   STATUS    RESTARTS   AGE
slurm-sample-d-0-0-45trk    1/1     Running   0          4m27s    # this is the daemon (slurmdbd)
slurm-sample-db-0-0-6jqkz   1/1     Running   0          4m27s    # this is that maria database
slurm-sample-s-0-0-xj5zr    1/1     Running   0          4m27s    # this is the login node (slurmctrl)
slurm-sample-w-0-0-8xtvw    1/1     Running   0          4m27s    # this is worker 0
slurm-sample-w-0-1-f25rp    1/1     Running   0          4m27s    # this is worker 1
```

You'll first want to see the database connect successfully:

```bash
$ kubectl logs slurm-sample-d-0-0-45trk -f
```
```console
slurmdbd: debug2: StorageType       = accounting_storage/mysql
slurmdbd: debug2: StorageUser       = slurm
slurmdbd: debug2: TCPTimeout        = 2
slurmdbd: debug2: TrackWCKey        = 0
slurmdbd: debug2: TrackSlurmctldDown= 0
slurmdbd: debug2: accounting_storage/as_mysql: acct_storage_p_get_connection: acct_storage_p_get_connection: request new connection 1
slurmdbd: debug2: Attempting to connect to slurm-sample-db-0-0.slurm-svc.slurm-operator.svc.cluster.local:3306
slurmdbd: slurmdbd version 21.08.6 started
slurmdbd: debug2: running rollup at Fri Jun 09 04:14:37 2023
slurmdbd: debug2: accounting_storage/as_mysql: as_mysql_roll_usage: Everything rolled up
slurmdbd: debug:  REQUEST_PERSIST_INIT: CLUSTER:linux VERSION:9472 UID:0 IP:10.244.0.152 CONN:7
slurmdbd: debug2: accounting_storage/as_mysql: acct_storage_p_get_connection: acct_storage_p_get_connection: request new connection 1
slurmdbd: debug2: Attempting to connect to slurm-sample-db-0-0.slurm-svc.slurm-operator.svc.cluster.local:3306
slurmdbd: debug2: DBD_FINI: CLOSE:0 COMMIT:0
slurmdbd: debug2: DBD_GET_CLUSTERS: called in CONN 7
slurmdbd: debug2: DBD_ADD_CLUSTERS: called in CONN 7
slurmdbd: dropping key time_start_end from table "linux_step_table"
slurmdbd: debug2: DBD_FINI: CLOSE:0 COMMIT:1
slurmdbd: debug2: DBD_FINI: CLOSE:1 COMMIT:0
```

And then watch the login node, which is starting the controller, registering the cluster, and starting again.
It normally would happen via a node reboot but we instead run it in a loop (and it seems to work). Note that I
found this step a bit slow.

```bash
kubectl logs -n slurm-operaslurm-sample-s-0-0-xj5zr -f
```
```bash
Hello, I am a server with slurm-sample-s-0-0.slurm-svc.slurm-operator.svc.cluster.local
Sleeping waiting for database...
---> Starting the MUNGE Authentication service (munged) ...
---> Sleeping for slurmdbd to become active before starting slurmctld ...
---> Starting the Slurm Controller Daemon (slurmctld) ...
 Adding Cluster(s)
  Name           = linux
slurmctld: debug:  slurmctld log levels: stderr=debug2 logfile=debug2 syslog=quiet
slurmctld: debug:  Log file re-opened
...
```
You'll see a lot of output stream to this log when it's finally running.

```console
slurmctld: debug2: Spawning RPC agent for msg_type REQUEST_PING
slurmctld: debug2: Tree head got back 0 looking for 2
slurmctld: debug2: Tree head got back 1
slurmctld: debug2: Tree head got back 2
slurmctld: debug2: node_did_resp slurm-sample-w-0-0.slurm-svc.slurm-operator.svc.cluster.local
slurmctld: debug2: node_did_resp slurm-sample-w-0-1.slurm-svc.slurm-operator.svc.cluster.local
slurmctld: debug:  sched/backfill: _attempt_backfill: beginning
slurmctld: debug:  sched/backfill: _attempt_backfill: no jobs to backfill
slurmctld: debug2: Testing job time limits and checkpoints
slurmctld: debug:  sched/backfill: _attempt_backfill: beginning
slurmctld: debug:  sched/backfill: _attempt_backfill: no jobs to backfill
slurmctld: debug2: Testing job time limits and checkpoints
slurmctld: debug2: Performing purge of old job records
slurmctld: debug:  sched: Running job scheduler for full queue.
slurmctld: debug2: Testing job time limits and checkpoints
```

Once you've verified the controller is running, you can shell into the control login node, and run sinfo or try a job:

```bash
kubectl exec -it slurm-sample-s-0-0-xj5zr bash
```
```bash
$ sinfo
```
```console
PARTITION AVAIL  TIMELIMIT  NODES  STATE NODELIST
normal*      up 5-00:00:00      2   idle slurm-sample-w-0-0.slurm-svc.slurm-operator.svc.cluster.local,slurm-sample-w-0-1.slurm-svc.slurm-operator.svc.cluster.local
```

Try running a "job" !

```bash
srun -N 2 hostname
slurm-sample-w-0-0.slurm-svc.slurm-operator.svc.cluster.local
slurm-sample-w-0-1.slurm-svc.slurm-operator.svc.cluster.local
```

When you are done, cleanup.

```bash
kind delete cluster
```
