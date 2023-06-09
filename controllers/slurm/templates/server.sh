#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# Default entrypoint with slurmctld, this is like a login node
echo "---> Starting the MUNGE Authentication service (munged) ..."
gosu munge /usr/sbin/munged

echo "---> Sleeping for slurmdbd to become active before starting slurmctld ..."

# A bit of a hack for now, need to check that slurmdbd is enabled
sleep 30
echo "---> Starting the Slurm Controller Daemon (slurmctld) ..."

until /usr/bin/sacctmgr --immediate add cluster name={{.Spec.ClusterName}}
do
    echo "Cluster {{.Spec.ClusterName}} is not ready... sleeping"
    sleep 5
done


while true
do
   exec gosu slurm /usr/sbin/slurmctld -i -Dvvv
done

{{template "exit" .}}
