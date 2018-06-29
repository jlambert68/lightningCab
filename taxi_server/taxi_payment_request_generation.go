package main

import (
	"jlambert/lightningCab/taxi_server/lightningConnection"
	taxiHW_stream_api "jlambert/lightningCab/taxi_hardware_servers/taxi_hardware_server_stream/taxi_hardware_grpc_stream_api" //"jlambert/lightningCab/taxi_hardware_server/taxi_hardware_grpc_api"
)

type amountStructure struct {
	timestamp          int64
	timeAmount         int64
	speedAmount        int64
	accelerationAMount int64
}

var lastAmount amountStructure
var lastPowerMessaurment taxiHW_stream_api.PowerStatusResponse

func gennerateInvoice() {

	invoice, err := lightningConnection.CreateInvoice("Payment Request for Taxi", 100, 180)
	if err != nil || invoice.Invoice == "" {
		logMessagesWithError(4, "Error when creating Invoice: ", err)

	} else {
		logMessagesWithOutError(4, "Invoice Created: ")

	}
}

func customerPaysPaymentRequest(check bool) (err error) {
	err = taxi.TaxiPaysPaymentRequest(check)
	return err
}

func calculateInvoiceAmount() {

}
