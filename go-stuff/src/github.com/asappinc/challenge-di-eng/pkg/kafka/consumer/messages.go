package consumer

import (
	"time"
)

// KafkaMessageInfo simple container for kafka message metadata
type KafkaMessageInfo struct {
	Topic     string
	Partition int32
	Offset    int64
	Timestamp time.Time
	Headers   map[string]string
	Key       []byte
}

// ConsumeMessage simple Kafka info and kafka message object
type ConsumeMessage struct {
	KafkaMessageInfo
	KafkaMessage interface{}
}

// New create a new ConsumeMessage
func NewMessage(info KafkaMessageInfo, message interface{}) *ConsumeMessage {
	c := new(ConsumeMessage)
	c.KafkaMessageInfo = info
	c.KafkaMessage = message
	return c
}

// ErrorMessage simple Kafka info and kafka message object
type ErrorMessage struct {
	KafkaMessageInfo
	Err error
}

// New create a new ConsumeMessage
func NewErrorMessage(info KafkaMessageInfo, err error) *ErrorMessage {
	c := new(ErrorMessage)
	c.KafkaMessageInfo = info
	c.Err = err
	return c
}

type NotificationMessage struct {
	Current  map[string][]int32
	Released map[string][]int32
	Claimed  map[string][]int32
	NoteType string
}
