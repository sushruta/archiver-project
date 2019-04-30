#!/bin/bash

sbt clean compile assembly

ls target/scala-2.11
exec flink run -c inc.asapp.flink.jobs.Archiver target/scala-2.11/archiver-0.1.jar
