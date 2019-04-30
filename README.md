# DI Project Eng/Platform:

Message buses serve as a key component for asynchronous systems as well as data pipelines.

In terms of data processing, we need to be able to store, archive, and query this data.

The key goal for this project is to create a data archiver for a message bus.  

By archiver we mean pulling data from the message bus, and save that data to disk.

### Final goals:

- Create a Kafka Archiver that consumes data from arbitrary topics and places them in neat
  files on the file system
  
  - Each "archive" should be partitioned by Message type (there are 3 different ones here)
    - meaning if the kafka messages are of type `object.one` then it should live in a folder
     `/path/to/archive/type=object.one/...`
     
  - Each "archive" should also be date partitioned 
    - meaning if the kafka message timestamp is 2019-08-01 03:23:03 the archive location 
    would be a folder of the name `/path/to/archive/dt=2019-08-01/hr=03/my_archive_file`)
  
  - The archive files themselves should not be a "file per message", but a collection of messages.
  
  - There are many edge cases and interesting error conditions for this, you will not have time to implement
  all the required "cruft" to be an ideal production ready system, so just make note of those places where you know
  there will be issues, and implement what you can
  
  - Along those same lines, make note of improvements, features, etc you would make if you had the time.
   
  - Place your code as a "command" in `./src/github.com/asappinc/archiver/cmd/archiver.go` 
  
  - It should be easily run via `go run archiver/cmd/archiver.go`
  
  
### Extra Credit: 

Create a Kafka Archiver that consumes data from arbitrary topics and places messages into a mysql or elasticsearch
so they can be queried.

### Some General notes about the system:

- There can be different messages on a given kafka topic: each message type should have it's own archive
- Message types are not limited to a single kafka topic
- Messages will be in the JSON format


### What you are provided:

- In Docker you will see a docker-compose.yml file that has a few applications running
	- Kafka/Zookeeper: a simple 1 node kafka "cluster"
	- Generator: a simple generator of data producing data into that kafka
	- MySQL: the RDS for one of your goals.
	- consumer: a base skeleton of consumer code to create your archiver from
		note: GoLang was chosen here, feel free to choose a language that suits you better
		This consumer is a simple "echo message" consumer (just emits messages to stdout)
    - producer: a simple 1 message per second producer

## Getting started

This will startup the relevant DBs, and drop you into a shell where go and go-path and configs are preset for you

1. Build things and get a dev shell
    
    
    ```
    docker-compose build archiver
    docker-compose run --rm dev bash
    ```


1. Start a simple producers in another shell

    ```
    docker-compose run --rm producer
    ```

1. Start a simple echo consumer in another shell

    ```
    docker-compose run --rm echoconsumer
    ```
    

You should see messages being produced from the producer

## Changes I made --

* I moved `src` and `Dockerfile` to `go` directory to place all the golang stuff there
* I created a `flinkapp` directory to house my flink code that will do the archiving
* I have appropriately changed `docker-compose.yml` to reflect the change in `Dockerfile` location
