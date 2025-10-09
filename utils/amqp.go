package util

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Error on dialing amqp %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error on creating channel %w", err)
	}

	return ch
}
