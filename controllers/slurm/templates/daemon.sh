#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

chmod +x /usr/local/bin/docker-entrypoint.sh
/usr/local/bin/docker-entrypoint.sh slurmdbd

{{template "exit" .}}