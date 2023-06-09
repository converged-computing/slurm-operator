#!/bin/sh

# Shared components for the broker and worker template
{{define "init"}}

# Copy over configs on all nodes
if [[ -f "/etc/slurm_operator/slurmdbd.conf" ]]; then 
    mkdir -p /etc/slurm
    cp /etc/slurm_operator/* /etc/slurm
    chown slurm /etc/slurm/slurmdbd.conf
fi

# Extra time for the database to setup
echo "Sleeping waiting for database..."
sleep 15

# Initialization commands
{{ .Node.Commands.Init}} > /dev/null 2>&1

# The working directory should be set by the CRD or the container
workdir=${PWD}

# And if we are using fusefs / object storage, ensure we can see contents
mkdir -p ${workdir}

# End init logic
{{end}}

{{define "exit"}}
sleep infinity
{{ if .Spec.Interactive }}sleep infinity{{ end }}
{{ end }}