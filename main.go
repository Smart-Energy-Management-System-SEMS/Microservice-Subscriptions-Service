package main

import (
	"log"

	"microservice-subscriptions-service/subscriptions"
)

func main() {
	if err := subscriptions.Start(); err != nil {
		log.Fatal(err)
	}
}
