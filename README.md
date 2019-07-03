[![Build Status](https://travis-ci.org/TomWright/deliver.svg?branch=master)](https://travis-ci.org/TomWright/deliver)
[![codecov](https://codecov.io/gh/TomWright/deliver/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/deliver)
[![Documentation](https://godoc.org/github.com/TomWright/deliver?status.svg)](https://godoc.org/github.com/TomWright/deliver)

# deliver

```
go get -u github.com/tomwright/deliver
```

Publish + consume messages with standard interfaces.

# Quick Start

This quick start assumes you already have your `Subscriber` or `Producer` created already.

For information on creating those objects, [see Implementations below](#implementations).

## Message

Below is an example message that may be published when a user has been created.

```
type UserCreated struct {
	Username string `json:"username"`
}

func (m *UserCreated) Type() string {
	return "user.created"
}

func (m *UserCreated) Payload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *UserCreated) WithPayload(bytes []byte) error {
	return json.Unmarshal(bytes, m)
}
```

## Publishing

- [Example Kafka Publisher](/examples/kafka_publisher)

## Subscribing

- [Example Kafka Subscriber](/examples/kafka_subscriber)

# Implementations

## Kafka

- `Message.Type()` response is used as the topic.
- Messages are marked before the given `ConsumerFn` is executed. 

Setup:
```
brokers := []string{"cmg-local-kafka:9092"}
publisher, err := deliver.NewKafkaPublisher(brokers)
subscriber := deliver.NewKafkaSubscriber(brokers)
```
