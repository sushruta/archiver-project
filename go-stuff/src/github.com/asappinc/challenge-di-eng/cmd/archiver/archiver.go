package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/asappinc/challenge-di-eng/pkg/kafka/consumer"
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

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *topics == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *group == "" {
		*group = "archiver-" + fmt.Sprint(time.Now().Unix())
	}

	c, err := consumer.NewConsumer(strings.Split(*brokers, ","), strings.Split(*topics, ","), *group)
	if err != nil {
		panic(err)
	}

	log.Println("Starting consumers ...")
	err = c.Connect()
	if err != nil {
		panic(err)
	}

	notesChan := make(chan *consumer.NotificationMessage)
	msgChan := make(chan *consumer.ConsumeMessage)
	errChan := make(chan *consumer.ErrorMessage)
	closer := make(chan struct{})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go c.Consume(notesChan, msgChan, errChan, closer)
	for {
		select {
		case n := <-notesChan:
			fmt.Println("Notification: ", n.NoteType)
		case e := <-errChan:
			fmt.Println("Error: ", e.Err.Error())
		case m := <-msgChan:
			// TODO: write somewhere!
			fmt.Printf(">> Topic: %s -- Message: %s\n<><><><><><><><><>\n", m.Topic, m.KafkaMessage)
		case <-signals:
			fmt.Println("shutting down archiver")
			close(closer)
			c.Close()
		}
	}
}
