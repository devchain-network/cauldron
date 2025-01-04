package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/IBM/sarama"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	brokers := []string{"127.0.0.1:9094"}
	topic := "deneme"

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return fmt.Errorf("consumer error: [%w]", err)
	}

	defer func() { _ = consumer.Close() }()

	partitionConsumer, errpc := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if errpc != nil {
		return fmt.Errorf("partition consumer error: [%w]", errpc)
	}
	defer func() { _ = partitionConsumer.Close() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	log.Printf("consuming messages from topic: [%s]\n", topic)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() { wg.Done() }()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					log.Printf(
						"key: [%s], value: [%s], offset: [%d] partition: [%d]\n",
						string(msg.Key), string(msg.Value), msg.Offset, msg.Partition,
					)
				}
			case err := <-partitionConsumer.Errors():
				log.Printf("error: %v\n", err)
			case <-signals:
				return
			}
		}
	}()
	wg.Wait()
	log.Println("return")

	return nil
}
