package main

import (
	"time"
	"github.com/sirupsen/logrus"

	"os"
	"log"
	"jlambert/lightningCab/common_config"
)



func (taxi *Taxi) InitLogger(filename string) {
	taxi.logger = logrus.StandardLogger()

	switch common_config.LoggingLevel {

	case logrus.DebugLevel:
		if common_config.MilliSecondsBetweenPaymentRequest < 1000 {
			log.Println("When using 'logrus.DebugLevel' minimum time between payment request mus be 1000 millisecond, curren value is: ", common_config.MilliSecondsBetweenPaymentRequest)
			os.Exit(0)
		}

	case logrus.InfoLevel:
		if common_config.MilliSecondsBetweenPaymentRequest < 1000 {
			log.Println("When using 'logrus.DebugLevel' minimum time between payment request mus be 1000 millisecond, curren value is: ", common_config.MilliSecondsBetweenPaymentRequest)
			os.Exit(0)
		}

	case logrus.WarnLevel:
		log.Println("'common_config.LoggingLevel': ", common_config.LoggingLevel)
		log.Println("'common_config.MilliSecondsBetweenPaymentRequest': ", common_config.MilliSecondsBetweenPaymentRequest)

	default:
		log.Println("Not correct value for debugging-level, this was used: ", common_config.LoggingLevel)
		os.Exit(0)

	}
	logrus.SetLevel(common_config.LoggingLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
		DisableSorting:  true,
	})


	//If no file then set standard out

	if filename == "" {
		taxi.logger.Out = os.Stdout

	} else {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			taxi.logger.Out = file
		} else {
			log.Println("Failed to log to file, using default stderr")
		}
	}

	// Should only be done from init functions
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(logger.Out, logger.Out, logger.Out))

}
