#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

# TODO something in here...

{{template "exit" .}}