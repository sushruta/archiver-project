#!/bin/bash

set -ex

service ssh restart

exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"
