package main

import (
	"flag"
	"github.com/Shopify/sarama"
	"github.com/asappinc/challenge-di-eng/pkg/objects"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	brokers = flag.String("brokers", os.Getenv("KAFKA_BROKERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose = flag.Bool("verbose", false, "Turn on logging")
	topics  = flag.String("topics", os.Getenv("KAFKA_TOPICS"), "kafka topics to produce to, as a comma separated list")
)

type KProducer struct {
	producer  sarama.AsyncProducer
	topicList []string
}

func (kf *KProducer) GetProducer(brokerList []string) sarama.AsyncProducer {

	if kf.producer != nil {
		return kf.producer
	}

	// For the access log, we are looking for AP semantics, with high throughput.
	// By creating batches of compressed messages, we reduce network I/O at a cost of more latency.
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start producer:", err)
	}

	// just dump to stdout for these
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write message:", err)
		}
	}()
	kf.producer = producer
	return kf.producer
}

func (kf *KProducer) Produce() {
	kfMessage := objects.FromObject(objects.RandomObject())
	tpc := kf.topicList[rand.Intn(len(kf.topicList))]
	kf.producer.Input() <- &sarama.ProducerMessage{
		Topic: tpc,
		Value: kfMessage,
	}
	log.Println("produced message to", tpc)
}

func main() {
	flag.Parse()

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[kafka-producer] ", log.LstdFlags)
	}

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *topics == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	prod := new(KProducer)
	prod.topicList = strings.Split(*topics, ",")
	prod.GetProducer(strings.Split(*brokers, ","))

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			prod.Produce()
		case <-signals:
			log.Println("stopping production: shutting down...")
			prod.producer.Close()
			break
		}
	}
	log.Println("shutdown complete")
}
