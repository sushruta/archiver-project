#!/bin/bash

$HADOOP_PREFIX/etc/hadoop/hadoop-env.sh

rm /tmp/*.pid

/usr/sbin/sshd
$HADOOP_PREFIX/sbin/start-dfs.sh
$HADOOP_PREFIX/sbin/start-yarn.sh
$HADOOP_PREFIX/sbin/mr-jobhistory-daemon.sh start historyserver

if [[ $1 == "-d" ]]; then
  while true; do sleep 1000; done
fi
