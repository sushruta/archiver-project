package main

import "fmt"
import "github.com/Shopify/sarama"
import "os"
import "os/signal"
import "time"
import "log"

type Batcher struct {
	from chan string
	to chan []string
	buffer []string
	filename string
}

func (b *Batcher) initialize(filename string) {
	// would initialize stuff here
	b.from = make(chan string, 1)
	b.to = make(chan []string, 1)
	b.buffer = make([]string, 0)
}

func (b *Batcher) accumulate(message string) {
	b.from <- message
}

func (b *Batcher) flush() {
	b.to <- b.buffer
	b.buffer = make([]string, 0)
	tmpbuf := <-b.to
	for _, message := range tmpbuf {
		log.Println("flushed message>>", message)
	}
}

func consume(topic string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	partitions, _ := master.Partitions(topic)
	consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		log.Printf("Topic %v Partitions: %v", topic, partitions)
		panic(err)
	}

	fmt.Println(" Start consuming topic ", topic)
	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case consumerError := <-consumer.Errors():
				errors <- consumerError
				log.Println("consumerError: ", consumerError.Err)
			case msg := <-consumer.Messages():
				consumers <- msg
				log.Println("Got message on topic ", topic, msg.Value)
			}
		}
	}(topic, consumer)

	return consumers, errors
}

func main() {
	config := sarama.NewConfig()
	config.ClientID = "go-archiver"
	config.Consumer.Return.Errors = true

	brokers := []string{"kafka:9092"}

	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	batcher := Batcher{}
	batcher.initialize("/tmp/go-dump")

	consumer, errors := consume("first_topic", master)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				go batcher.accumulate(string(msg.Value))
			case consumerError := <-errors:
				log.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				log.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	timer := time.NewTimer(2 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				log.Println("timed out ... ")
				if len(batcher.buffer) > 0 {
					log.Println("flushing out the buffer")
					batcher.flush()
				}
				timer = time.NewTimer(2 * time.Second)
			default:
				batcher.buffer = append(batcher.buffer, <-batcher.from)
				if len(batcher.buffer) == 4 {
					log.Println("capacity reached ... flushing the buffer")
					batcher.flush()
					stop := timer.Stop()
					if stop {
						timer = time.NewTimer(2 * time.Second)
					}
				}
			}
		}
	}()

	<-doneCh
	fmt.Println("completed. exiting ...")
}
