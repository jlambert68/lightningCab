package main

import (
	"jlambert/lightningCab/taxi_server/lightningConnection"
	taxiHW_stream_api "github.com/jlambert68/lightningCab/grpc_api/taxi_hardware_grpc_stream_api" //"jlambert/lightningCab/taxi_hardware_server/taxi_hardware_grpc_api"
	"io"
	"golang.org/x/net/context"
	"math"
	"jlambert/lightningCab/common_config"
	"errors"
	"github.com/sirupsen/logrus"
)

type amountStructure struct {
	timestamp          int64
	timeAmount         int64
	speedAmount        int64
	accelerationAmount int64
}

//var lastAmountToPay int64
//var lastAmountStructure amountStructure
//var lastPowerMessaurment taxiHW_stream_api.PowerStatusResponse
var paymentRequestIsPaid bool = false
var abortPaymentRequestGeneration bool = false
var firstMissedPaymentTimeOut bool = false
var lastMissedPaymentTimeOut bool = false

type lastPaymentData_struct struct {
	lastAmountToPay_satoshi int64
	lastAmountToPay_sek     float32
	lastPowerMessaurment    taxiHW_stream_api.PowerStatusResponse
	lastReceivedAmountdata  amountStructure
}

var lastPaymentData lastPaymentData_struct

func generateInvoice() (lightningConnection.PendingInvoice, error) {

	var err error = nil
	var invoice lightningConnection.PendingInvoice

	// Don't create invoice of zero amount
	if lastPaymentData.lastAmountToPay_satoshi != 0 {
		//paymentRequestIsPaid = false
		invoice, err = lightningConnection.CreateInvoice("Payment Request for Taxi", lastPaymentData.lastAmountToPay_satoshi, 360)
		if err != nil || invoice.Invoice == "" {
			logMessagesWithError(4, "Error when creating Invoice: ", err)

		} else {
			logMessagesWithOutError(4, "Invoice Created: " + invoice.Invoice)

		}
	} else {
		logMessagesWithOutError(4, "Amount is zero, so no Invoice is generated: ")
		err = errors.New("Amount is zero, so no Invoice is generated")
	}

	//log.Println("Sending back the following data!")
	//spew.Println(invoice)
	return invoice, err
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
			taxi.logger.Warning("State machine is not in correct State to be able to accept payments")
		}
	}

	return err
}

func calculateInvoiceAmount() {
	/*
	log.Println("Underlag för Invoice:::::::")
	log.Println(lastPaymentData.lastPowerMessaurment.GetSpeed())
	log.Println()
	log.Println()
	*/

	lastPaymentData.lastReceivedAmountdata.speedAmount = int64(math.Round(float64(lastPaymentData.lastPowerMessaurment.GetSpeed()) * common_config.SpeedSatoshiPerSecond / 100.0 * common_config.MilliSecondsBetweenPaymentRequest / 1000))
	lastPaymentData.lastReceivedAmountdata.accelerationAmount = int64(math.Round(float64(lastPaymentData.lastPowerMessaurment.GetAcceleration()) * common_config.MaxAccelarationSatoshiPerSecond / 100.0 * common_config.MilliSecondsBetweenPaymentRequest / 1000))
	lastPaymentData.lastReceivedAmountdata.timeAmount = int64(math.Round(float64(common_config.SpeedSatoshiPerSecond * common_config.MilliSecondsBetweenPaymentRequest / 1000)))

	lastPaymentData.lastAmountToPay_satoshi = lastPaymentData.lastReceivedAmountdata.speedAmount + lastPaymentData.lastReceivedAmountdata.accelerationAmount + lastPaymentData.lastReceivedAmountdata.timeAmount
	lastPaymentData.lastAmountToPay_sek = float32(lastPaymentData.lastAmountToPay_satoshi) * common_config.BTCSEK / common_config.SatoshisPerBTC

	//spew.Println(lastPaymentData)
	//log.Println(lastPaymentData)
}

func receiveEnginePowerdata(client taxiHW_stream_api.TaxiStreamHardwareClient, messasurePowerMessage *taxiHW_stream_api.MessasurePowerMessage) {
	//log.Println("Starting Engine Powerdata stream %v", messasurePowerMessage)
	taxi.logger.WithFields(logrus.Fields{
		"messasurePowerMessage":    messasurePowerMessage,
	}).Info("Starting Engine Powerdata stream")

	ctx := context.Background()
	//	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	//	defer cancel()

	stream, err := client.MessasurePowerConsumption(ctx, messasurePowerMessage)
	if err != nil {
		//log.Fatalf("Problem to connect to Taxi Engine Stream: ", client, err)
		taxi.logger.WithFields(logrus.Fields{
			"client":    client,
			"error":	err,
		}).Fatal("Problem to connect to Taxi Engine Stream")
	}
	for {
		powerMessaurement, err := stream.Recv()
		if err == io.EOF {
			//log.Println("HMMM, skumt borde inte slutat h" +
			//	"är när vi tar emot EngineStream, borde avsluta Taxi-server")
			taxi.logger.Error("HMMM, skumt borde inte slutat h" +
				"är när vi tar emot EngineStream, borde avsluta Taxi-server")

			break
		}
		if err != nil {
			//log.Fatalf("Problem when streaming from Taxi Engine Stream:", client, err)
			taxi.logger.WithFields(logrus.Fields{
				"client":    client,
				"error":	err,
			}).Fatal("Problem when streaming from Taxi Engine Stream")
		}
		lastPaymentData.lastPowerMessaurment = *powerMessaurement

		calculateInvoiceAmount()
		//log.Println("New Data for invoice is calculated")

	}
}
