package main

import (
	"time"
	"github.com/sirupsen/logrus"

	"os"
	"log"
	"jlambert/lightningCab/common_config"
)



func (customer *Customer) InitLogger(filename string) {
	customer.logger = logrus.StandardLogger()
	if common_config.MilliSecondsBetweenPaymentRequest < 1000 {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(common_config.LoggingLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
		DisableSorting:  true,
	})


	//If no file then set standard out

	if filename == "" {
		customer.logger.Out = os.Stdout

	} else {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			customer.logger.Out = file
		} else {
			log.Println("Failed to log to file, using default stderr")
		}
	}

	// Should only be done from init functions
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(logger.Out, logger.Out, logger.Out))

}
