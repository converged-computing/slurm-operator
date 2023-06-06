#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# TODO something in here...

{{template "exit" .}}
