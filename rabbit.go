package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love the world!")
}

func rabbitRun(chn *amqp.Channel) {
	for {
		err := chn.Publish(
			"logs", // exchange
			"",     // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("ping"),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent ping")
		time.Sleep(5000 * time.Millisecond)
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

	chn, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer chn.Close()

	_, err = chn.QueueDelete(
		"logs",
		false,
		false,
		false,
	)
	failOnError(err, "Failed to delete old exchange")

	err = chn.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	go rabbitRun(chn)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
