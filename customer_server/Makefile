# cat -e -t -v Makefile

.DEFAULT_GOAL := build

build:
	@echo "go build..."
	go build customer.go logrus_logger.go customer_stream_api.go

run:
	./customer
