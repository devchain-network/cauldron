package main

import (
	"log"

	"github.com/devchain-network/cauldron/internal/kafkaconsumer"
)

func main() {
	if err := kafkaconsumer.Run(); err != nil {
		log.Fatal(err)
	}
}
