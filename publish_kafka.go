package deliver

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"time"
)

func NewKafkaPublisher(brokerList []string) (Publisher, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V1_1_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Timeout = time.Second
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}
	p := &kafkaPublisher{
		producer: producer,
	}
	return p, nil
}

type kafkaPublisher struct {
	producer sarama.SyncProducer
}

func (x *kafkaPublisher) Publish(m Message) error {
	if x.producer == nil {
		return errors.New("cannot publish on closed publisher")
	}

	payload, err := m.Payload()
	if err != nil {
		return fmt.Errorf("could not get message payload: %s", err)
	}
	_, _, err = x.producer.SendMessage(&sarama.ProducerMessage{
		Topic: m.Type(),
		Value: sarama.StringEncoder(payload),
	})

	if err != nil {
		return err
	}

	return nil
}

func (x *kafkaPublisher) Close() error {
	if x.producer == nil {
		return errors.New("missing producer cannot be closed")
	}
	return x.producer.Close()
}
