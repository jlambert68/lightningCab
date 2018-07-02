package main

import (
	"log"
	"fmt"
	"time"
	//"google.golang.org/grpc"
	"golang.org/x/net/context"
	taxi_api "jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"jlambert/lightningCab/common_config"
)

/*
rpc AskTaxiForPrice (Enviroment) returns (Price) {
}

//Accept price from Taxi
rpc AcceptPrice (Enviroment) returns (AckNackResponse) {
}

// Returns a stream of PaymentRequests to Customer
rpc PaymentRequestStream (Enviroment) returns (stream PaymentRequest) {
}
*/

// Customer connects and ask for price
func (s *taxiServiceServer) AskTaxiForPrice(ctx context.Context, environment *taxi_api.Enviroment) (*taxi_api.Price, error) {

	log.Println("Incoming: 'AskTaxiForPrice'")
	fmt.Println("sleeping...for 3 seconds")
	time.Sleep(3 * time.Second)

	currentPrice := &taxi_api.Price{
		true,
		"",
		33000,
		16000,
		10000,
		taxi_api.TimeUnit_SecondsBetweenPaymentmentRequests,
		1,
		taxi_api.PriceUnit_SatoshiPerSecond}

	// Check if State machine accepts State change
	err := taxi.CustomerConnects(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case taxi_api.TestOrProdEnviroment_Test:
			// Customer Connects to Taxi
			log.Println("Customer Connects to Taxi:")
			currentPrice.Acknack = true
			currentPrice.Comments = "Welcome to the best Taxi you can find around here!"

		case taxi_api.TestOrProdEnviroment_Production:
			// Customer Connects to Taxi
			log.Println("Customer Connects to Taxi:")
			currentPrice.Acknack = true
			currentPrice.Comments = "Welcome to the best Taxi you can find around here!"

		default:
			logMessagesWithOutError(4, "Unknown incomming enviroment: "+environment.TestOrProduction.String())
			currentPrice.Acknack = false
			currentPrice.Comments = "Unknown incomming enviroment: " + environment.TestOrProduction.String()
		}

		err = taxi.CustomerConnects(false)
		if err != nil {
			logMessagesWithError(4, "There was a problem to change state in Taxi State machine: ", err)
			currentPrice.Acknack = false
			currentPrice.Comments = "There was a problem to change state in Taxi State machine"
		}

	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able have customer connects to it: ", err)
		currentPrice.Acknack = false
		currentPrice.Comments = "State machine is not in correct state to be able have customer connects to it"
	}

	return currentPrice, nil

}

// Customer accepts price
func (s *taxiServiceServer) AcceptPrice(ctx context.Context, environment *taxi_api.Enviroment) (*taxi_api.AckNackResponse, error) {

	log.Println("Incoming: 'AcceptPrice'")
	fmt.Println("sleeping...for 3 seconds")
	time.Sleep(3 * time.Second)

	var acknack bool
	var returnMessage string

	// Check if State machine accepts State change
	err := taxi.CustomerAcceptsPrice(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case taxi_api.TestOrProdEnviroment_Test:
			// Customer Connects to Taxi
			log.Println("Customer Connects to Taxi:")
			acknack = true
			returnMessage = "Welcome to the best Taxi you can find around here!"

		case taxi_api.TestOrProdEnviroment_Production:
			// Customer Connects to Taxi
			log.Println("Customer Connects to Taxi:")
			acknack = true
			returnMessage = "Welcome to the best Taxi you can find around here!"

		default:
			logMessagesWithOutError(4, "Unknown incomming enviroment: "+environment.TestOrProduction.String())
			acknack = false
			returnMessage = "Unknown incomming enviroment: " + environment.TestOrProduction.String()
		}

		err = taxi.CustomerAcceptsPrice(false)
		if err != nil {
			logMessagesWithError(4, "There was a problem to change state in Taxi State machine: ", err)
			acknack = false
			returnMessage = "There was a problem to change state in Taxi State machine"
		}

	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able have customer connects to it: ", err)
		acknack = false
		returnMessage = "State machine is not in correct state to be able have customer connects to it"
	}

	return &taxi_api.AckNackResponse{acknack, returnMessage}, nil
}

func (s *taxiServiceServer) PaymentRequestStream(enviroment *taxi_api.Enviroment, stream taxi_api.Taxi_PaymentRequestStreamServer) (err error) {
	log.Printf("Incoming: 'PaymentRequestStream'")

	err = nil
	paymentRequestResponse := &taxi_api.PaymentRequest{
		""}

	for {
		paymentRequestResponse.LightningPaymentRequest, err = generateInvoice()
		if err := stream.Send(paymentRequestResponse); err != nil {
			return err
			log.Printf("Error when streaming back: 'PaymentRequestStream'")
			break
		}

		time.Sleep(common_config.MilliSecondsBetweenPaymentRequest * time.Millisecond)
		if paymentRequestIsPaid == false {
			log.Printf("PaymentRequest not paid in time. Stop Taxi and wait for payment")
			//Stop Stream for a while and send state machine in wait mode
			taxi.PaymentsStopsComing(false)
		}

	}
	return nil
}

