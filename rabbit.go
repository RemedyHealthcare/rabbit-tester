package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Hub struct {
	Conn *amqp.Connection
	Chn  *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love the world!")
}

func rabbitRun(h *Hub) {
	for {
		err := h.Chn.Publish(
			"logs", // exchange
			"",     // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("ping"),
			})
		if err != nil {
			h.Chn, err = h.Conn.Channel()
			failOnError(err, "Failed to open a channel")
			err = h.Chn.Publish(
				"logs", // exchange
				"",     // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte("ping"),
				})
			failOnError(err, "Failed to publish even after recovering channel")
			if err != nil {
				log.Printf(" [x] SUCCESSFUL CHANNEL RECOVERY")
				log.Printf(" [x] Sent ping")
			}
		} else {
			log.Printf(" [x] Sent ping")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	rabbitURL := os.Getenv("rabbitURL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	var hub Hub
	hub.Conn = conn

	hub.Chn, err = hub.Conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer hub.Chn.Close()

	_, err = hub.Chn.QueueDelete(
		"logs",
		false,
		false,
		false,
	)
	failOnError(err, "Failed to delete old exchange")

	err = hub.Chn.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	go rabbitRun(&hub)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
