# cat -e -t -v Makefile

.DEFAULT_GOAL := build

build:
	@echo "go build..."
	go build taxi_server.go taxi_payment_request_generation.go taxi_server_api.go logrus_logger.go

run:
	./taxi_server


buildRasPi:
	@echo "go build for Rasberry Pi..."
	env GOOS=linux GOARCH=arm GOARM=7 go build -o  taxi_server_raspi taxi_server.go taxi_server_api.go taxi_payment_request_generation.go logrus_logger.go

move:
	scp taxi_server_raspi jlambert@192.168.2.51:/home/jlambert/go_workspace/src/jlambert/lightningCab/taxi_server