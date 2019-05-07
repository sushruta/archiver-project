package main

import (
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/asappinc/challenge-di-eng/pkg/kafka/consumer"
	//"github.com/asappinc/challenge-di-eng/pkg/objects"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
	// "encoding/json"
	"bufio"
)

type Router struct {
	baseDir     string
	batchers    map[string]*Batcher
	batcherChan chan *Batcher
}

func NewRouter(baseDir string) *Router {
	var router Router
	router.baseDir = baseDir
	router.batchers = make(map[string]*Batcher)
	router.batcherChan = make(chan *Batcher)

	return &router
}

func (r *Router) Collect(message *consumer.ConsumeMessage) {
	batcher := r.GetOrCreateBatcher(message)
	jsonstr := string(message.KafkaMessage.([]uint8))
	// jsonstr := JsonString(message.KafkaMessage.(*objects.KafkaMessage))
	batcher.Accumulate(jsonstr)
}

func (r *Router) GetOrCreateBatcher(message *consumer.ConsumeMessage) *Batcher {
	topic := message.KafkaMessageInfo.Topic
	timestamp := message.KafkaMessageInfo.Timestamp
	year, month, day := timestamp.Date()
	hour := timestamp.Hour()

	completePath := fmt.Sprintf("%s/%s/dt=%04d-%02d-%02d/hr=%02d", r.baseDir, topic, year, month, day, hour)
	batcher, ok := r.batchers[completePath]
	if !ok {
		batcher = NewBatcher(completePath)
		r.batchers[completePath] = batcher
		r.batcherChan <- batcher
	}
	return batcher
}

type Batcher struct {
	from chan string
	to chan []string
	buffer []string

	completePath string
	currentFileNumber int
	currentFilename string
	currentFile *os.File
	writer *bufio.Writer
	numLinesWritten int
}

func NewBatcher(cp string) *Batcher {
	var b Batcher
	b.from = make(chan string, 1)
	b.to = make(chan []string, 1)
	b.buffer = make([]string, 0)

	b.completePath = cp
	b.currentFileNumber = 0

	os.MkdirAll(b.completePath, 0777)
	b.currentFilename = fmt.Sprintf("%s/part-%04d", b.completePath, b.currentFileNumber)
	b.currentFile, b.writer = CreateFileAndWriter(b.currentFilename)

	b.numLinesWritten = 0
	return &b
}

func (b *Batcher) Collect() {
	timer := time.NewTimer(2 * time.Minute)
	for {
		select {
		case <-timer.C:
			if b.numLinesWritten >= 128 {
				b.numLinesWritten = 0
				b.currentFileNumber = b.currentFileNumber + 1
				b.writer.Flush()
				b.currentFile.Close()

				b.currentFilename = fmt.Sprintf("%s/part-%04d", b.completePath, b.currentFileNumber)
				b.currentFile, b.writer = CreateFileAndWriter(b.currentFilename)
			}
			timer = time.NewTimer(2 * time.Minute)
		default:
			b.buffer = append(b.buffer, <-b.from)
			if len(b.buffer) == 4 {
				b.Flush()
				stop := timer.Stop()
				if stop {
					timer = time.NewTimer(2 * time.Minute)
				}
			}
		}
	}
}

func CreateFileAndWriter(filename string) (*os.File, *bufio.Writer) {
	fmt.Println("creating a file:", filename)
	flw, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(flw)
	log.Println("created a file for writing:", filename)

	return flw, w
}

func (b *Batcher) Accumulate(message string) {
	b.from <- message
}

func (b *Batcher) Flush() {
	// decide on the file to open and write to it
	b.to <- b.buffer
	b.buffer = make([]string, 0)
	tmpbuf := <-b.to
	for _, message := range tmpbuf {
		b.writer.WriteString(fmt.Sprintf("%s\n", message))
	}
	b.numLinesWritten = b.numLinesWritten + len(tmpbuf)
}

/**
func JsonString(msg *objects.KafkaMessage) string {
	var obj interface{}
	switch msg.Type {
	case "object.one":
		obj = (msg.Obj).(*objects.ObjectOne)
	case "object.two":
		obj = (msg.Obj).(*objects.ObjectTwo)
	case "object.three":
		obj = (msg.Obj).(*objects.ObjectThree)
	}

	jsonstr, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return string(jsonstr)
}
**/

// Flags

var (
	brokers = flag.String("brokers", os.Getenv("KAFKA_BROKERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose = flag.Bool("verbose", false, "Turn on logging")
	topics  = flag.String("topics", os.Getenv("KAFKA_TOPICS"), "kafka topics to produce to, as a comma separated list")
	group   = flag.String("group", os.Getenv("KAFKA_CONSUMER_GROUP"), "consumer group")
	baseDir   = flag.String("baseDir", os.Getenv("BASE_DIR"), "base directory for data persistence")
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
		*group = "go-archiver-" + fmt.Sprint(time.Now().Unix())
	}

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[echo-consumer] ", log.LstdFlags)
	}

	if *baseDir == "" {
		*baseDir = "/tmp/go-archiver-datalake"
	}

	router := NewRouter(*baseDir)

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

	doneCh := make(chan struct{})

	go c.Consume(notesChan, msgChan, errChan, closer)
	go func() {
		for {
			select {
			case n := <-notesChan:
				fmt.Println("Notification: ", n.NoteType)
			case e := <-errChan:
				fmt.Println("ERROR: ", e.Err.Error())
			case m := <-msgChan:
				router.Collect(m)
				c.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "")
			case <-signals:
				fmt.Println("Got shutdown, stopping consumer")
				close(closer)
				c.Close()
				doneCh <- struct{}{}
			}
		}
	}()

	go func() {
		for {
			b := <-router.batcherChan
			b.Collect()
		}
	}()

	<-doneCh
}

