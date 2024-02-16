runcounter:
	go run .\counter

runclient:
	go run .\client

runevent:
	go run .\event

runorder:
	go run .\order

runall:
	go run .\all


installs:
	go get github.com/confluentinc/confluent-kafka-go/v2/kafka
	go get github.com/google/uuid