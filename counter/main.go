package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go-kafka-event/event"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Counter struct {
	*http.Server
	Producer     *kafka.Producer
	deliveryChan chan kafka.Event
}

func NewCounter(addr string, Producer *kafka.Producer) *Counter {
	c := &Counter{
		Producer:     Producer,
		deliveryChan: make(chan kafka.Event),
	}

	c.Server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(c.takeOrder),
	}

	return c
}

func (c *Counter) takeOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.handleOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *Counter) handleOrder(w http.ResponseWriter, r *http.Request) {
	amount := r.URL.Query().Get("amount")
	if amount == "" {
		http.Error(w, "Missing amount", http.StatusBadRequest)
		return
	}

	numAmount, err := strconv.Atoi(amount)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	if numAmount <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	order := event.NewOrder(numAmount)

	val, err := order.MarshalBinary()
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}

	err = c.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &event.OrderProcessedTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(order.OrderID),
		Value: val,
	}, c.deliveryChan)
	if err != nil {
		http.Error(w, "Failed to produce order", http.StatusInternalServerError)
		return
	}

	e := <-c.deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		http.Error(w, "Failed to deliver order", http.StatusInternalServerError)
		return
	} else {
		log.Printf(w, "Produced message to popic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		w.WriteHeader(http.StatusCreated)
		w.Headerr().Add("Content-Type", "application/json")
		w.Write(val)
	}
}

func main() {
	producer, errr := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		panic(err)
	}

	log.Println("Counter is ready to take orders.")

	stopChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	counter := NewCounter(":8080", producer)

	go func() {
		defer close(stopChan)
		defer close(sigChan)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = counter.Shutdown(ctx)
	}()

	log.Println("Starting counter...")

	var err = counter.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	<-stopChan

	log.Println("Counter is shutting down")
}
