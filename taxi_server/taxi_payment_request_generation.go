package main

import (
	"jlambert/lightningCab/taxi_server/lightningConnection"
	taxiHW_stream_api "jlambert/lightningCab/taxi_hardware_servers/taxi_hardware_server_stream/taxi_hardware_grpc_stream_api" //"jlambert/lightningCab/taxi_hardware_server/taxi_hardware_grpc_api"
	"log"
	"time"
	"io"
	"golang.org/x/net/context"
	"math"
	"jlambert/lightningCab/common_config"
)

type amountStructure struct {
	timestamp          int64
	timeAmount         int64
	speedAmount        int64
	accelerationAMount int64
}

var lastAmountToPay int64
var lastAmountStructure amountStructure
var lastPowerMessaurment taxiHW_stream_api.PowerStatusResponse
var paymentRequestIsPaid bool = false

func generateInvoice() (string, error) {

	var invoice lightningConnection.PendingInvoice

	paymentRequestIsPaid = false
	invoice, err := lightningConnection.CreateInvoice("Payment Request for Taxi", lastAmountToPay, 180)
	if err != nil || invoice.Invoice == "" {
		logMessagesWithError(4, "Error when creating Invoice: ", err)

	} else {
		logMessagesWithOutError(4, "Invoice Created: ")

	}

	return invoice.Invoice, err
}

func customerPaysPaymentRequest(check bool) (err error) {
	paymentRequestIsPaid = true
	err = nil

	// If statemachine is in state: 'StateTaxiIsReadyToDrive' everthing is normal
	if !taxi.TaxiStateMachine.IsInState(StateTaxiIsReadyToDrive) {

		//Check if Taxi state is in State: StateTaxiIsWaitingForPayment
		if taxi.TaxiStateMachine.IsInState(StateTaxiIsWaitingForPayment) {
			// Send Statemachine to Continue streaming PaymentRequests
			err = taxi.continueStreamingPaymentRequests(false)
		} else {
			// Inform Customer that Invoice must be paid in x seconds otherwise Taxi is ready for new customer
			log.Println("State machine is not in correct State to be able to accept payments")
		}
	}

	return err
}

func calculateInvoiceAmount() {
	lastAmountStructure.speedAmount = int64(math.Round(float64(lastPowerMessaurment.GetSpeed()) * common_config.SpeedSatoshiPerSecond / 100.0))
	lastAmountStructure.accelerationAMount = int64(math.Round(float64(lastPowerMessaurment.GetSpeed()) * common_config.MaxAccelarationSatoshiPerSecond / 100.0))
	lastAmountStructure.timeAmount = int64(math.Round(float64(lastPowerMessaurment.GetSpeed()) * common_config.SpeedSatoshiPerSecond / 100.0))

	lastAmountToPay = lastAmountStructure.speedAmount + lastAmountStructure.accelerationAMount + lastAmountStructure.timeAmount
}

func receiveEnginePowerdata(client taxiHW_stream_api.TaxiStreamHardwareClient, messasurePowerMessage *taxiHW_stream_api.MessasurePowerMessage) {
	log.Printf("Starting Engine Powerdata stream %v", messasurePowerMessage)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.MessasurePowerConsumption(ctx, messasurePowerMessage)
	if err != nil {
		log.Fatalf("Problem to connect to Taxi Engine Stream: ", client, err)
	}
	for {
		powerMessaurement, err := stream.Recv()
		if err == io.EOF {
			log.Println("HMMM, skumt borde inte slutat här när vi tar emot EngineStream, borde avsluta Taxi-server")
			break
		}
		if err != nil {
			log.Fatalf("Problem when streaming from Taxi Engine Stream:", client, err)
		}
		lastPowerMessaurment = *powerMessaurement

		calculateInvoiceAmount()

	}
}
