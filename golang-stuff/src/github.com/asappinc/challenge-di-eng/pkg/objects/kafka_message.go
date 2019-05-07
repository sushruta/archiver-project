package objects

import (
	"encoding/json"
)

type KafkaMessage struct {
	Obj  BaseObject
	Type string

	encoded []byte
	err     error
}

func (km *KafkaMessage) encode() {
	if km.encoded == nil && km.err == nil {
		km.encoded, km.err = json.Marshal(km)
	}
}

func (km *KafkaMessage) Length() int {
	km.encode()
	return len(km.encoded)
}

func (km *KafkaMessage) Encode() ([]byte, error) {
	km.encode()
	return km.encoded, km.err
}

func FromObject(obj BaseObject) *KafkaMessage {
	return &KafkaMessage{
		Obj:  obj,
		Type: obj.Type(),
	}
}
