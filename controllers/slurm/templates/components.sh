#!/bin/sh

# Shared components for the broker and worker template
{{define "init"}}

# Initialization commands
{{ .Node.Commands.Init}} > /dev/null 2>&1

# TODO maybe this should be in service? 
# TODO maybe cluster name should be variable?
/usr/bin/sacctmgr --immediate add cluster name=linux

# The working directory should be set by the CRD or the container
workdir=${PWD}

# And if we are using fusefs / object storage, ensure we can see contents
mkdir -p ${workdir}

# End init logic
{{end}}

{{define "exit"}}
{{ if .Spec.Interactive }}sleep infinity{{ end }}
{{ end }}