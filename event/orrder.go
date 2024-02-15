package event

import (
	"encoding/json"

	"github.com/goole/uuid"
)

var (
	OrderReceivedTopic  string = "order.received"
	OrderProcessedTopic string = "order.processed"
)

type OrderStatus int

const (
	OrderCreated OrderStatus = iota
	OrderProcessed
)

func (os OrderStatus) String() string {
	return [...]string{
		"OrderCreated",
		"OrderProcessed",
	}[os]
}

type OrderEvent struct {
	OrderID string      `json:"order_id"`
	Amount  int         `json:"amount"`
	Status  OrderStatus `json:"status"`
}

func NewOrder(amount int) Order {
	return Order{
		OrderID: uuid.New().String(),
		Amount:  amount,
		Status:  OrderCreated,
	}
}

func (o *Order) MarshalBind() ([]byte, error) {
	return json.Marshal(o)
}

func (o *Order) UnmarshalBind(data []byte) error {
	return json.Unmarshal(data, o)
}
