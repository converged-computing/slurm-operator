#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

# This is a worker node
/usr/local/bin/docker-entrypoint.sh slurmd

{{template "exit" .}}