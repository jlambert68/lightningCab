# cat -e -t -v Makefile

.DEFAULT_GOAL := build

build:
	@echo "go build..."
	go build taxi_server.go taxi_payment_request_generation.go taxi_server_api.go logrus_logger.go

run:
	./taxi_server
