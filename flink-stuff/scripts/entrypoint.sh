#!/bin/bash

sbt clean compile assembly

exec flink run -c inc.asapp.flink.apps.Archiver target/scala-2.11/archiver-assembly-0.1.jar
