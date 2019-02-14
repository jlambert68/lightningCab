// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"fmt"
	"github.com/jlambert68/lightningCab/grpc_api/proto/server"
	"jlambert/lightningCab/customer_server/webmain"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)


// ******************************************************************************
// Ask Taxi for Price
func CallBackAskTaxiForPrice(emptyParameter *server.EmptyParameter) (*server.Price_UI, error) {
	log.Println("Incoming: 'AskTaxiForPrice'")

	returnMessage := &server.Price_UI{
		Acknack:                   true,
		Comments: "Hej Hopp, from 'CallBackAskTaxiForPrice'",
		SpeedAmountSatoshi:        0,
		AccelerationAmountSatoshi: 1,
		TimeAmountSatoshi:         2,
		SpeedAmountSek:            3,
		AccelerationAmountSek:     4,
		TimeAmountSek:             5,
		Timeunit:                  0,
		PaymentRequestInterval:    0,
		Priceunit:                 0,
	}


	return returnMessage, nil
}





// ******************************************************************************
// Customer accepts Price

func CallBackAcceptPrice(emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error){

	log.Println("Incoming: 'AcceptPrice'")

	returnMessage := &server.AckNackResponse{
		Acknack:  true,
		Comments: "Hej Hopp, from 'CallBackAcceptPrice'",
	}


	return returnMessage, nil
}



// ******************************************************************************
// Customer Halts Payments
func CallBackHaltPayments(haltPaymentRequestMessage *server.HaltPaymentRequest) (*server.AckNackResponse, error){

	log.Println("Incoming: 'HaltPayments' with parameter: ", haltPaymentRequestMessage.Haltpayment)

	returnMessage := &server.AckNackResponse{
		Acknack:  true,
		Comments: "Hej Hopp, from 'CallBackHaltPayments'",
	}


	return returnMessage, nil
}



// ******************************************************************************
// Customer Leaves Taxi

func CallBackLeaveTaxi(emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error){

	log.Println("Incoming: 'LeaveTaxi'")

	returnMessage := &server.AckNackResponse{
		Acknack:  true,
		Comments: "Hej Hopp, from 'CallBackLeaveTaxi'",
	}

	return returnMessage, nil
}




func main(){

	go webmain.Webmain(CallBackAskTaxiForPrice, CallBackAcceptPrice, CallBackHaltPayments, CallBackLeaveTaxi)

	// Set system in wait mode for externa input ...
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		os.Exit(0)
	}()

	for {
		fmt.Println("sleeping...for another 5 minutes")
		time.Sleep(300 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}

}
