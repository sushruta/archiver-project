#!/bin/bash

set -ex

service ssh restart

if [[ $1 == "taskmanager" ]]; then
  while true; do sleep 1000; done
fi

if [[ $1 == "jobmanager" ]]; then
  start-cluster.sh
  while true; do sleep 1000; done
fi
