package main

import (
	"log"

	"github.com/devchain-network/cauldron/internal/kafkacp/kafkaconsumer/kafkagithubconsumer"
)

func main() {
	if err := kafkagithubconsumer.Run(); err != nil {
		log.Fatal(err)
	}
}
