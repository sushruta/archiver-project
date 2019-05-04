package main

import (
	"bufio"
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

type Batcher struct {
	from     chan string
	to       chan []string
	buffer   []string
	writer   *bufio.Writer
	flw      *os.File
	filename string
}

func (b *Batcher) initialize(filename string) {
	b.from = make(chan string, 1)
	b.to = make(chan []string, 1)
	b.buffer = make([]string, 16)
	b.filename = filename
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	b.flw = f
	b.writer = bufio.NewWriter(b.flw)
}

func (b *Batcher) accumulatefn(message string) {
	b.from <- message
}

func (b *Batcher) flushfn() {
	b.to <- b.buffer
	b.buffer = make([]string, 0)
	tmpbuf := <-b.to
	for _, message := range tmpbuf {
		b.writer.WriteString(message)
	}
	b.writer.Flush()
}

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

	if *baseDir == "" {
		*baseDir = "/tmp/kafka.dump"
	}

	log.Println("done with reading all the env vars")
	log.Println("base dir -- ", *baseDir)
	log.Println("kafka topics -- ", *topics)
	log.Println("base dir -- ", *group)
	log.Println("base dir -- ", *brokers)

	c, err := consumer.NewConsumer(strings.Split(*brokers, ","), strings.Split(*topics, ","), *group)
	if err != nil {
		panic(err)
	}

	log.Println("able to init a consumer")

	batcher := Batcher{}

	log.Println("Starting consumers ...")
	err = c.Connect()
	if err != nil {
		panic(err)
	}

	log.Println("no problem starting the consumer ...")

	notesChan := make(chan *consumer.NotificationMessage)
	msgChan := make(chan *consumer.ConsumeMessage)
	errChan := make(chan *consumer.ErrorMessage)
	closer := make(chan struct{})

	batcher.initialize(*baseDir)
	log.Println("initialized the batcher")
	defer batcher.flw.Close()
	log.Println("deferred the batcher")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go c.Consume(notesChan, msgChan, errChan, closer)
	log.Println("starting the go routine for messages")

	timer := time.NewTimer(2 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				log.Println("timed out ... ")
				if len(batcher.buffer) > 0 {
					log.Println("flushing out the buffer")
					batcher.flushfn()
				}
				timer = time.NewTimer(2 * time.Second)
			default:
				batcher.buffer = append(batcher.buffer, <-batcher.from)
				if len(batcher.buffer) == 5 {
					batcher.flushfn()
					stop := timer.Stop()
					if stop {
						timer = time.NewTimer(2 * time.Second)
					}
				}
			}
		}
	}()

	for {
		select {
		case n := <-notesChan:
			fmt.Println("Notification: ", n.NoteType)
		case e := <-errChan:
			fmt.Println("Error: ", e.Err.Error())
		case m := <-msgChan:
			go batcher.accumulatefn(fmt.Sprintf("%s", m.KafkaMessage))
			c.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "") // is this correct?
		case <-signals:
			fmt.Println("shutting down archiver")
			close(closer)
			c.Close()
		}
	}
}
