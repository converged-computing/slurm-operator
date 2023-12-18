# Federated SLURM

I guess this is a few SLURM clusters together, connected with a common database?
Create the cluster:

## 1. Create Cluster

```bash
kind create cluster
```

You'll need to install the jobset API, which eventually will be added to Kubernetes proper (but is not yet!)

```bash
VERSION=v0.2.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```

From the root, install the slurm operator:

```bash
make test-deploy-recreate
```

See logs for the operator

```bash
kubectl logs -n slurm-operator-system slurm-operator-controller-manager-6f6945579-9pknp
```

## 2. Create Shared Database

Create the separate slurm database first. Note that we have set the hostname to match what the job expects (not a fully qualified domain name, but the `setHostnameAsFQDN` combined with the service will ensure it is). We have also set the username and password for slurm.

```bash
kubectl apply -f database.yaml
```

## 3. Create Slurm Blue 🔵️

Next, create the first interactive cluster. The database is disabled, meaning it will not be created (but still look for one, the one we've created). We also have provided the fully qualified domain name of the database.

```bash
kubectl apply -f slurm-0.yaml 
```

Wait until all of the containers are running:

```bash
kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
slurm-blue-d-0-0-kfzz4   1/1     Running   0          12s  # this is the daemon (slurmdbd)
slurm-blue-s-0-0-zz6bm   1/1     Running   0          12s  # this is the login node (slurmctrl)
slurm-blue-w-0-0-ljdvt   1/1     Running   0          12s  # this is worker 0
slurm-blue-w-0-1-tlkc6   1/1     Running   0          12s  # this is worker 1
slurm-db                 1/1     Running   0          51s  # this is our federated (shared) database
```

You'll first want to see the database connect successfully from the daemon:

```bash
kubectl logs slurm-blue-d-0-0-hdbdr -f
```
```console
slurmdbd: debug2: StorageType       = accounting_storage/mysql
slurmdbd: debug2: StorageUser       = slurm
slurmdbd: debug2: TCPTimeout        = 2
slurmdbd: debug2: TrackWCKey        = 0
slurmdbd: debug2: TrackSlurmctldDown= 0
slurmdbd: debug2: accounting_storage/as_mysql: acct_storage_p_get_connection: acct_storage_p_get_connection: request new connection 1
slurmdbd: debug2: Attempting to connect to slurm-db.slurm-svc.slurm-operator.svc.cluster.local:3306
slurmdbd: slurmdbd version 21.08.6 started
slurmdbd: debug2: running rollup at Fri Jun 09 04:14:37 2023
slurmdbd: debug2: accounting_storage/as_mysql: as_mysql_roll_usage: Everything rolled up
slurmdbd: debug:  REQUEST_PERSIST_INIT: CLUSTER:linux VERSION:9472 UID:0 IP:10.244.0.152 CONN:7
slurmdbd: debug2: accounting_storage/as_mysql: acct_storage_p_get_connection: acct_storage_p_get_connection: request new connection 1
slurmdbd: debug2: Attempting to connect to slurm-db.slurm-svc.slurm-operator.svc.cluster.local:3306
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
kubectl logs slurm-blue-s-0-0-xj5zr -f
```
```bash
Hello, I am a server with slurm-blue-s-0-0.slurm-svc.slurm-operator.svc.cluster.local
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
slurmctld: debug:  sched/backfill: _attempt_backfill: no jobs to backfill
slurmctld: debug2: Testing job time limits and checkpoints
slurmctld: debug2: Testing job time limits and checkpoints
slurmctld: debug2: Performing purge of old job records
slurmctld: debug:  sched: Running job scheduler for full queue.
slurmctld: debug2: Testing job time limits and checkpoints
slurmctld: debug:  Spawning ping agent for slurm-blue-w-0-0.slurm-svc.default.svc.cluster.local,slurm-blue-w-0-1.slurm-svc.default.svc.cluster.local
slurmctld: debug2: Spawning RPC agent for msg_type REQUEST_PING
slurmctld: debug2: Tree head got back 0 looking for 2
slurmctld: debug2: Tree head got back 1
slurmctld: debug2: Tree head got back 2
slurmctld: debug2: node_did_resp slurm-blue-w-0-0.slurm-svc.default.svc.cluster.local
slurmctld: debug2: node_did_resp slurm-blue-w-0-1.slurm-svc.default.svc.cluster.local
slurmctld: debug:  sched/backfill: _attempt_backfill: beginning
slurmctld: debug:  sched/backfill: _attempt_backfill: no jobs to backfill
```

Once you've verified the controller is running, you can shell into the control login node, and run sinfo or try a job:

```bash
kubectl exec -it slurm-blue-s-0-0-xj5zr bash
```
```bash
sinfo
```
```console
PARTITION AVAIL  TIMELIMIT  NODES  STATE NODELIST
normal*      up 5-00:00:00      2   idle slurm-blue-w-0-0.slurm-svc.default.svc.cluster.local,slurm-blue-w-0-1.slurm-svc.default.svc.cluster.local
```

Try running a "job" !

```bash
srun -N 2 hostname
slurm-blue-w-0-0.slurm-svc.default.svc.cluster.local
slurm-blue-w-0-1.slurm-svc.default.svc.cluster.local
```

## 4. Create Slurm Pink 🟤️

Now let's do the same, but create slurm pink! This will be a second slurm cluster that is using the same database.
I'm not sure what is going to happen.

```bash
kubectl apply -f slurm-1.yaml
```
You can again look at the daemon (d) and login node (s) and see they also are connecting to the same database,
and you can submit jobs. You now have two slurm clusters running on the same database.

```console
NAME                     READY   STATUS    RESTARTS   AGE
slurm-blue-d-0-0-nbv92   1/1     Running   0          7m13s
slurm-blue-s-0-0-rt7lj   1/1     Running   0          7m13s
slurm-blue-w-0-0-8747b   1/1     Running   0          7m13s
slurm-blue-w-0-1-lv9c9   1/1     Running   0          7m13s
slurm-db                 1/1     Running   0          7m17s
slurm-pink-d-0-0-m99mp   1/1     Running   0          2m32s
slurm-pink-s-0-0-whgzn   1/1     Running   0          2m32s
slurm-pink-w-0-0-69zfx   1/1     Running   0          2m32s
slurm-pink-w-0-1-n6ggh   1/1     Running   0          2m32s
```

TODO next: test the random federation commands and play around!

## 5. Cleanup

When you are done, cleanup.

```bash
$ kind delete cluster
```
