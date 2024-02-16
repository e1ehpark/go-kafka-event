package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go-kafka-event/event"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	c := flag.Int("c", 1, "number of clients")
	flag.Parse()

	log.Printf("Starting %d clients", *c)

	consumer := newConsumer()
	defer consumer.Close()

	err := consumer.SubscribeTopics([]string{"event.OrderProcessedTopic"}, nil)
	if err != nil {
		panic(err)
	}

	stopChan := make(chan struct{})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		for {
			select {
			case <-sigChan:
				close(stopChan)
				return
			default:
				msg, err := consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					continue
				}

				order := event.Order{}

				err = order.UnmarshalBinary(msg.Value)
				if err != nil {
					log.Printf("Failed to unmarshal order: %s", err)
					continue
				}

				log.Printf("Order ID s% is ready for pickup. Enjoy your coffee!", order.OrderID)
			}
		}
	}()

	client := http.DefaultClient

	for i := 0; i < *c; i++ {
		go func() {
			amount := rand.Intn(5) + 1
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/order?amount=%d", amount), nil)
			if err != nil {
				panic(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				panic("order not accepted")
			}

			var order event.Order

			err = json.NewDecoder(resp.Body).Decode(&order)
			if err != nil {
				panic(err)
			}

			log.Printf("Order accepted: %s", order.OrderID)

		}()

		time.Sleep(1 * time.Second)
	}

	<-stopChan
}

func newConsumer() *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	return consumer
}
