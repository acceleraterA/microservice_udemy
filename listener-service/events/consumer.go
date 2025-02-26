package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// receiving event from queue
type Consumer struct {
	Conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		Conn: conn,
	}
	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (consumer *Consumer) setup() error {
	// create a channel
	channel, err := consumer.Conn.Channel()
	if err != nil {
		return err
	}
	return declareExchange(channel)
	// declare the queue
	// consume the message
}

// push the message to the queue
type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	// create a channel
	ch, err := consumer.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}
	for _, s := range topics {
		// bind s to the queue
		ch.QueueBind(
			q.Name,       // queue name
			s,            // routing key
			"logs_topic", // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return err
		}
	}
	message, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}
	//
	forever := make(chan bool)
	go func() {
		for d := range message {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			go handlePayload(payload)
		}
	}()
	fmt.Printf("waiting for msg on [exchange,queue] [logs_topic,%s]\n", q.Name)
	<-forever // read from the forever channel and block until a value is received,
	//since no value is sent to the channel, the program will block indefinitely
	return nil
}

// handle payload from queue for log or auth
func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		//log whatever we get
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		//send to auth service

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}

}
func logEvent(entry Payload) error {
	// create json we'll send to the log service
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	// call the log service
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close() // close the body when we're done
	//make sure get back correct status code
	if response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}
