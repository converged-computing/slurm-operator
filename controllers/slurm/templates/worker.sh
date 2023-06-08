#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

# This is a worker node
chmod +x /usr/local/bin/docker-entrypoint.sh
/usr/local/bin/docker-entrypoint.sh slurmd

{{template "exit" .}}