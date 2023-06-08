#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# Default entrypoint with slurmctld, this is like a login node
chmod +x /usr/local/bin/docker-entrypoint.sh
/usr/local/bin/docker-entrypoint.sh slurmctld

{{template "exit" .}}
