#cat -e -t -v Makefile

.DEFAULT_GOAL := build

build:
	@echo "go build..."
	go build taxi_hardware_server.go

buildRasPi:
	@echo "go build for Rasberry Pi..."
	env GOOS=linux GOARCH=arm GOARM=7 go build -o  taxi_hardware_server_raspi taxi_hardware_server.go

run:
	./taxi_hardware_server

move:
	scp taxi_hardware_server_raspi jlambert@192.168.1.93:/home/jlambert/go_workspace/src/jlambert/lightningCab/

