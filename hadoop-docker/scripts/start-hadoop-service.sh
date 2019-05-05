#!/bin/bash

set -ex

hdfs namenode -format
start-dfs.sh
