#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic
{{template "init" .}}

# This is a worker node
{{ template "munge" .}}

echo "---> Waiting for slurmctld to become active before starting slurmd..."
sleep 30
echo "---> Starting the Slurm Node Daemon (slurmd) ..."
exec /usr/sbin/slurmd -Dvvv

{{template "exit" .}}