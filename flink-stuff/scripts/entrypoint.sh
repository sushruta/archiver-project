#!/bin/bash

exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"

## sbt clean compile assembly
## 
## exec flink run -c com.asappinc.Archiver target/scala-2.11/archiver-assembly-0.1.jar
