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

// So, the way the code works is that there is a router in front of the system
// that receives the messages and decides to send it to a particular batcher.
// The batcher receives the message and acts on it.
// TODO - take the numbers out and put them in variables for easy changing
//        for example - batch size, timeout interval, number of batchers, etc.

// Router is the gateway to the messages in the system
// Based on the topic, dt and hr, it decides which batcher
// will get this message and sends the message to that batcher
// A router is fully described by --
//
// 1. the base dir which indicates the location of all the data
// 2. batchers map[string]*batchers -- a map of all the batchers that can receive messages
// 3. batcherChan An active sequence of batchers which will execute the collecting go routine
//
// TODO - old batchers needs to be removed so that the number of batchers in the memory doesn't
// keep increasing monotonically.
type Router struct {
	baseDir     string
	batchers    map[string]*Batcher
	batcherChan chan *Batcher
}

// A helper function to create a new router
// and initialize the data structures
func NewRouter(baseDir string) *Router {
	var router Router
	router.baseDir = baseDir
	router.batchers = make(map[string]*Batcher)
	router.batcherChan = make(chan *Batcher)

	log.Println("created a new router")

	return &router
}

// This method collects the messages and decides on where
// to send them. Given the topic and timestamp, it decides
// which batcher needs to get it. If the batcher doesn't
// exist, it will create it first.
func (r *Router) Collect(message *consumer.ConsumeMessage) {
	batcher := r.GetOrCreateBatcher(message)
	jsonstr := string(message.KafkaMessage.([]uint8))
	// jsonstr := JsonString(message.KafkaMessage.(*objects.KafkaMessage))
	batcher.Accumulate(jsonstr)
}

// Given a topic and timestamp, check if a batcher exists and return it.
// If it doesn't, create one and return it.
func (r *Router) GetOrCreateBatcher(message *consumer.ConsumeMessage) *Batcher {
	topic := message.KafkaMessageInfo.Topic
	timestamp := message.KafkaMessageInfo.Timestamp
	year, month, day := timestamp.Date()
	hour := timestamp.Hour()

	completePath := fmt.Sprintf("%s/%s/dt=%04d-%02d-%02d/hr=%02d", r.baseDir, topic, year, month, day, hour)
	batcher, ok := r.batchers[completePath]
	if !ok {
		log.Println("batcher for", completePath, " doesn't exist. Creating it ...")
		batcher = NewBatcher(completePath)
		r.batchers[completePath] = batcher
		r.batcherChan <- batcher
	}
	return batcher
}

// A Batcher is fully described by
// 1. from chan string - this is gateway for messages entering
// 2. buffer []string - this contains a tmp collection of messages
// 3. to chan[]string - this is the gateway for messages to be stored
//
// 4. completePath - tells us where the part files will be stored
// 5. currentFileNumber - the part file number currently in action
// ... the file descriptors for writing
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

// Create a NEW batcher. Every topic, dt and hr combination
// will have a unique batcher which will be responsible for
// collecting, persisting and rotating the files in that
// directory
//
// For Example -- for topic - first_topic, dt - 2019-05-07, hr - 13
// a unique batcher will exist which will get all the messages
// belonging to the above combination and persist them to `part-` files
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

	log.Println("created a new batcher", cp)
	return &b
}

// This method collects messages coming in a topic
// and adds them to buffer. We want to hold a maximum
// of 8 elements in a buffer before persisting them
// to the disk. That happens in the flush method
// Also, if the number of lines in the file is >= 32
// we would like to close it out and open a new file
func (b *Batcher) Collect() {
	timer := time.NewTimer(20 * time.Second)
	for {
		select {
		case <-timer.C:
			// file is sufficiently big - rotate it
			if b.numLinesWritten >= 32 {
				b.numLinesWritten = 0
				b.currentFileNumber = b.currentFileNumber + 1
				b.writer.Flush()
				b.currentFile.Close()

				b.currentFilename = fmt.Sprintf("%s/part-%04d", b.completePath, b.currentFileNumber)
				b.currentFile, b.writer = CreateFileAndWriter(b.currentFilename)
				log.Println("start writing to a new file", b.completePath)
			}
			timer = time.NewTimer(20 * time.Second)
		default:
			b.buffer = append(b.buffer, <-b.from)
			if len(b.buffer) == 8 {
				log.Println("reached capacity", b.completePath)
				b.Flush()
			}
		}
	}
}

// Given a filename, this function simply returns a file descriptor
// and an auxilliary bufio writer for better writing
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

// We don't want to incur wasteful CPU cycles
// persisting every line to the file separately.
// We accumulate the message in a buffer which
// on reaching a certain capacity and can be flushed
// out to the disk
func (b *Batcher) Accumulate(message string) {
	b.from <- message
}

// This method takes the contents of the buffer and writes
// them to the file. Also, it keeps track of the number of
// lines it has added to the file so far for bookkeeping purposes
func (b *Batcher) Flush() {
	// decide on the file to open and write to it
	log.Println("in the flush function", b.completePath)
	b.to <- b.buffer
	b.buffer = make([]string, 0)
	tmpbuf := <-b.to
	for _, message := range tmpbuf {
		b.writer.WriteString(fmt.Sprintf("%s\n", message))
	}
	b.writer.Flush()
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
				// send the message to the router
				go router.Collect(m)
				c.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "")
			case <-signals:
				fmt.Println("Got shutdown, stopping consumer")
				close(closer)
				c.Close()
				doneCh <- struct{}{}
			}
		}
	}()

	// start a go routine to let new batchers collect their events
	go func() {
		for {
			b := <-router.batcherChan
			go b.Collect()
		}
	}()

	<-doneCh
}

