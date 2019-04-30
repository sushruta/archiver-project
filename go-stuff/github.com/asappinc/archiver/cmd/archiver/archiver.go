package main

import (
	"flag"
	"os"
)

var (
	brokers = flag.String("brokers", os.Getenv("KAFKA_BROKERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose = flag.Bool("verbose", false, "Turn on logging")
	topics  = flag.String("topics", os.Getenv("KAFKA_TOPICS"), "kafka topics to produce to, as a comma separated list")
	group   = flag.String("group", os.Getenv("KAFKA_CONSUMER_GROUP"), "consumer group")
	baseDir = flag.String("todir", os.Getenv("ARCHIVE_BASE_LOCATION"), "root folder to place archived files")
)

func main() {
	flag.Parse()

	// Do all the things

}
