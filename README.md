#### Changes I Made

* In the produce.go and consumer.go files, I added `config.Version = sarama.V0.11.0.0` to make sure timestamps are embedded in the message and I can use them.
* created archiver.go file which is my attempt at writing an archiving tool for kafka messages
* I decreased the time between producing messages thereby making the stream a little *faster*

#### How to run

Just do a `docker-compose up --build`

I usually do a `docker-compose up --build | grep -v kafka_1 | grep -v zookeeper_1` to get rid of the verbose logs of kafka and zookeeper.

By default, the target location of the archive is `/tmp/go-archive-datalake` but it can be changed using the env var - `BASE_DIR`

To inspect the data dump, I was `exec-ing` into the docker container and checking the contents of the above directory. Some things I was checking was -

* the number of lines in each file
* the number of parts in each batch
* ... etc. ...


#### Challenges and Reflections

* Having no experience in writing code in Golang, it was definitely rewarding to see one write this application in golang.
* The code is not robust to very fast moving kafka streams - there may be one or two race conditions which I wanted to get rid of but didn't get the time. For example, if a part file is being closed while a new batch of messages comes in, where would the new file be stored. Depending on the newfile number, I may end up with a bit of data loss. I would ideally fix it by putting a channel or a mutex in between.
* The code is not fault tolerant - let us say the code crashes for some reason and we bring it back. All the messages which were sent in the meanwhile are lost. Basically, we don't keep track of the offset where we are. That can be fixed by constantly storing some metadata information.
* Some golang pain points - casting things from `interface{}` to the data type. Understanding some internals of sarama to make the timestamp working took more time than expected.

#### Attempt to put Flink into action

* Rather than write the entire archiver in its full glory from scratch, I made an attempt to use Apache Flink to do all this for me. However, to get all the features that are missing from my golang code - 

* fault tolerance
* checkpointing
* disaster recovery
* exactly one semantics

I would have required to deploy a full standalone flink cluster which is able to connect to an HDFS backend.

* I was able to create a `Hadoop` cluster (`1` namenode and `2` datanodes)
* A `Flink` cluster (`1` jobmanager and `2` taskmanagers).
* I was also able to connect the cluster to a `Zookeeper` quorom.
* I was able to submit my flink job to the cluster.

However, the cluster wasn't able to fully connect to HDFS backend completely and crashed. Given some more time, I would have loved to get this working and make sure a version closer to production grade archiving tool was ready. 
