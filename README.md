[![Build Status](https://travis-ci.org/TomWright/deliver.svg?branch=master)](https://travis-ci.org/TomWright/deliver)
[![codecov](https://codecov.io/gh/TomWright/deliver/branch/master/graph/badge.svg)](https://codecov.io/gh/TomWright/deliver)
[![Documentation](https://godoc.org/github.com/TomWright/deliver?status.svg)](https://godoc.org/github.com/TomWright/deliver)

# deliver

```
go get -u github.com/tomwright/deliver
```

Publish + consume messages with standard interfaces.

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