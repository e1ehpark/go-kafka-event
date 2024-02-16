runcounter:
	go run .\counter

runclient:
	go run .\client

runevent:
	go run .\event

runbarista:
	go run .\barista

runall:
	go run .\counter
	go run .\client
	go run .\event
	go run .\barista

installs:
	go get github.com/confluentinc/confluent-kafka-go/v2/kafka
	go get github.com/google/uuid