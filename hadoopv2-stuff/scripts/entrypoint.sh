#!/bin/bash

set -ex

# $HADOOP_CONF_DIR/hadoop-env.sh

hdfs namenode -format
start-dfs.sh

## # add some adhoc directories
## $HADOOP_PREFIX/bin/hdfs dfs -mkdir /user
## $HADOOP_PREFIX/bin/hdfs dfs -mkdir /user/sashi
## 
## $HADOOP_PREFIX/sbin/start-yarn.sh
## $HADOOP_PREFIX/sbin/mr-jobhistory-daemon.sh start historyserver

exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"

