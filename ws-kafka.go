package main

import (
	"fmt"

	"github.com/Shopify/sarama"
)

var (
	client           sarama.Client
	httpRequestTopic KafkaTopic
)

type KafkaTopic struct {
	Client     sarama.Client
	Topic      string
	Producer   sarama.AsyncProducer
	Producable chan<- *sarama.ProducerMessage
	Consumer   sarama.Consumer
	Consumable sarama.PartitionConsumer
}

func (q KafkaTopic) RegisterTopic(id string, callback func(key, value string)) {
	// Clients[id] =
}

func CreateKafkaTopic() *KafkaTopic {
	client, err := sarama.NewClient([]string{"kafka:9092"}, sarama.NewConfig())
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Kafka Client connected: %v\n", client)
	}

	topic := "http-request"
	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Kafka Producer connected: %v\n", producer)
	}
	producable := producer.Input()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Kafka Consumer connected: %v\n", consumer)
	}

	consumable, err := consumer.ConsumePartition(topic, 0, 0)
	if err != nil {
		panic(err)
	}

	return &KafkaTopic{client, topic, producer, producable, consumer, consumable}
}

func (q KafkaTopic) Start() {
	go func() {
		for {
			fmt.Printf("Producer Error: %v\n", <-q.Producer.Errors())
		}
	}()
	go func() {
		for {
			fmt.Printf("Producer Success: %v\n", <-q.Producer.Successes())
		}
	}()
}

func (q KafkaTopic) Close() {
	q.Consumer.Close()
	q.Producer.Close()
	q.Client.Close()
}

func (q KafkaTopic) Enqueue(key, value string) {
	q.Producable <- &sarama.ProducerMessage{
		Topic: q.Topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}
}

var offset = 0

func (q KafkaTopic) Dequeue() (string, string) {
	kafkaOut := <-q.Consumable.Messages()
	return string(kafkaOut.Key), string(kafkaOut.Value)
}
