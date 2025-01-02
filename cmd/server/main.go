package main

import (
	"log"

	"github.com/devchain-network/cauldron/internal/apiserver"
)

func main() {
	if err := apiserver.Run(); err != nil {
		log.Fatal(err)
	}
}
