#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

/usr/local/bin/docker-entrypoint.sh slurmdbd

{{template "exit" .}}