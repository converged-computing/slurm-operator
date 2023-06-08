#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

echo "---> Starting the MUNGE Authentication service (munged) ..."
gosu munge /usr/sbin/munged

echo "---> Starting the Slurm Database Daemon (slurmdbd) ..."

. /etc/slurm/slurmdbd.conf
until echo "SELECT 1" | mysql -h $StorageHost -u$StorageUser -p$StoragePass 2>&1 > /dev/null
do
    echo "-- Waiting for database to become active ..."
    sleep 2
done

echo "-- Database is now active ..."
exec gosu slurm /usr/sbin/slurmdbd -Dvvv

{{template "exit" .}}