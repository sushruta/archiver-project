package main

import (
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/asappinc/challenge-di-eng/pkg/kafka/consumer"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

// Flags

var (
	brokers = flag.String("brokers", os.Getenv("KAFKA_BROKERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose = flag.Bool("verbose", false, "Turn on logging")
	topics  = flag.String("topics", os.Getenv("KAFKA_TOPICS"), "kafka topics to produce to, as a comma separated list")
	group   = flag.String("group", os.Getenv("KAFKA_CONSUMER_GROUP"), "consumer group")
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
		*group = "echo-consumer-" + fmt.Sprint(time.Now().Unix())
	}

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[echo-consumer] ", log.LstdFlags)
	}

	c, err := consumer.NewConsumer(strings.Split(*brokers, ","), strings.Split(*topics, ","), *group)
	if err != nil {
		panic(err)
	}

	log.Println("Starting consumers ... ")
	err = c.Connect()
	if err != nil {
		panic(err)
	}

	notesChan := make(chan *consumer.NotificationMessage)
	msgChan := make(chan *consumer.ConsumeMessage)
	errChan := make(chan *consumer.ErrorMessage)
	closer := make(chan struct{})

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go c.Consume(notesChan, msgChan, errChan, closer)
	for {
		select {
		case n := <-notesChan:
			fmt.Println("Notification: ", n.NoteType)
		case e := <-errChan:
			fmt.Println("ERROR: ", e.Err.Error())
		case m := <-msgChan:
			fmt.Printf("Message: Topic: %s Partition: %d Offset: %d :: \n -------- \n %s \n -------- \n ",
				m.Topic, m.Partition, m.Offset, m.KafkaMessage,
			)
			c.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "")
		case <-signals:
			fmt.Println("Got shutdown, stopping consumer")
			close(closer)
			c.Close()
			return
		}
	}

}
