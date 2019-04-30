package consumer

import (
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"sync"
	"sync/atomic"
	"time"
)

var ErrorAlreadyConsuming = errors.New("consumer is already consuming")
var ErrorTopicsMustBeUnique = errors.New("topics must be unique")

// KfConsumer Simple consumer group object for kafaka
type KfConsumer struct {
	conf     *cluster.Config
	clientID string // kafka client ID

	// Client cluster client
	Client *cluster.Client

	// Consumer base consumers
	Consumer *cluster.Consumer

	// Broker strings
	Brokers []string

	// Topics to consume
	Topics []string

	// GroupID the consumer group ID string
	GroupID string

	lock          sync.RWMutex
	isConsuming   int32
	commitOnClose bool
}

// NewConsumer get a new consumer
func NewConsumer(brokers []string, topics []string, groupID string) (*KfConsumer, error) {
	s := new(KfConsumer)
	s.Brokers = brokers
	s.Topics = topics
	s.commitOnClose = true

	// make sure topic list is UNIQUE
	tMap := make(map[string]struct{})
	for _, t := range topics {
		if _, ok := tMap[t]; ok {
			return nil, ErrorTopicsMustBeUnique
		}
		tMap[t] = struct{}{}
	}

	s.GroupID = groupID
	s.clientID = "golang-echo-consumer"

	s.conf = cluster.NewConfig()
	s.conf.Consumer.Return.Errors = true
	s.conf.Group.Return.Notifications = true
	s.conf.Metadata.RefreshFrequency = 120 * time.Second
	s.conf.Metadata.Full = true
	s.conf.Consumer.Fetch.Min = 1
	s.conf.Consumer.Fetch.Max = 1000
	s.conf.ClientID = s.clientID
	s.conf.Consumer.MaxProcessingTime = time.Millisecond * 250
	s.conf.Consumer.Fetch.Default = 1024 * 1024
	s.conf.Metadata.RefreshFrequency = 120 * time.Second

	return s, nil
}

// Connect to kafka, grab the partitions
func (s *KfConsumer) Connect() (err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.Client, err = cluster.NewClient(s.Brokers, s.conf)
	if err != nil {
		return err
	}

	s.Consumer, err = cluster.NewConsumerFromClient(s.Client, s.GroupID, s.Topics)
	if err != nil {
		return err
	}
	return err
}

// TglCommitOnClose toggle the commit latest committed message on consumer shutdown
func (s *KfConsumer) TglCommitOnClose(commitOnClose bool) {
	s.commitOnClose = commitOnClose
}

// Consume use call backs to consume messages and errors instead of channels
func (s *KfConsumer) Consume(
	notificationChan chan *NotificationMessage,
	messageChan chan *ConsumeMessage,
	errorChan chan *ErrorMessage,
	closer chan struct{}) (err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	if atomic.LoadInt32(&s.isConsuming) > 0 {
		return ErrorAlreadyConsuming
	}

	atomic.StoreInt32(&s.isConsuming, 1)

	// wait for "rebalance ok" to start consuming
	for n := range s.Consumer.Notifications() {
		if n.Type == cluster.RebalanceOK {
			sarama.Logger.Println("rebalance ok, starting consumers")
			break
		}
	}
	go s.runConsume(notificationChan, messageChan, errorChan, closer)
	return err
}

func (s *KfConsumer) runConsume(
	notificationChan chan *NotificationMessage,
	messageChan chan *ConsumeMessage,
	errorChan chan *ErrorMessage,
	closer chan struct{}) {

	onConsumer := s.Consumer
	tClose := closer
	notifyCh := onConsumer.Notifications()
	errCh := onConsumer.Errors()
	msgCh := onConsumer.Messages()

	for {
		select {
		case n, more := <-notifyCh:
			if !more {
				notifyCh = nil
				continue
			}
			if n == nil {
				continue
			}
			notificationChan <- &NotificationMessage{
				Claimed:  n.Claimed,
				Released: n.Released,
				Current:  n.Current,
				NoteType: fmt.Sprint(n.Type),
			}

		case e, more := <-errCh:
			if !more {
				errCh = nil
				continue
			}
			kfInfo := KafkaMessageInfo{
				Topic:     "<unknonwn>",
				Offset:    -1,
				Partition: -1,
			}

			errorChan <- NewErrorMessage(kfInfo, e)
		case m, more := <-msgCh:
			if m == nil {
				msgCh = nil
				continue
			}
			if !more {
				continue
			}

			messageChan <- NewMessage(
				KafkaMessageInfo{
					Topic:     m.Topic,
					Partition: m.Partition,
					Offset:    m.Offset,
					Timestamp: m.Timestamp,
					Key:       m.Key,
				},
				m.Value,
			)

		case <-tClose:
			sarama.Logger.Println("stopping consumption")
			atomic.StoreInt32(&s.isConsuming, 0)
			return
		}
	}
}

// MarkPartitionOffset mark the offset topic, partition
func (s *KfConsumer) MarkPartitionOffset(topic string, partition int32, offset int64, metadata string) error {
	s.Consumer.MarkPartitionOffset(topic, partition, offset, metadata)
	return nil
}

// CommitOffsets force offset sync
func (s *KfConsumer) CommitOffsets() error {
	return s.Consumer.CommitOffsets()
}

// Close all partition consumers and main connect objects
func (s *KfConsumer) Close() (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sarama.Logger.Println("closing consumer")

	if s.Consumer != nil && s.commitOnClose {
		err = s.Consumer.CommitOffsets()
		if err != nil {
			sarama.Logger.Println("error on final commit offsets", "error", err)
		}
	}

	if s.Client != nil && !s.Client.Closed() {
		s.Consumer.Close()
		s.Client.Close()
	}
	atomic.StoreInt32(&s.isConsuming, 0)
	return err
}
