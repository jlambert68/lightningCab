package main

import (
	"log"
	"time"
	"jlambert/lightningCab/customer_server/customer_ui_stream_api"
)

func (s *Customer_UIServiceStreamServer) UIPriceAndStateStream(emptyParameter *customer_ui_stream_api.EmptyParameter, stream customer_ui_stream_api.CustomerUIPriceStream_UIPriceAndStateStreamServer) (err error) {
	log.Printf("Incoming: 'UIPriceAndStateStream'")

	err = nil

	for {
		priceAndStateRespons := &customer_ui_stream_api.UIPriceAndStateRespons{
			true,
			"",
			customer.lastReceivedInvoice.SpeedAmountSatoshi,
			customer.lastReceivedInvoice.AccelerationAmountSatoshi,
			customer.lastReceivedInvoice.TimeAmountSatoshi,
			customer.lastReceivedInvoice.SpeedAmountSek,
			customer.lastReceivedInvoice.AccelerationAmountSek,
			customer.lastReceivedInvoice.TimeAmountSek,
			customer.lastReceivedInvoice.TotalAmountSatoshi,
			customer.lastReceivedInvoice.TotalAmountSek,
			time.Now().UnixNano(),
			allowedgRPC_CustomerUI(),
		}

		if err := stream.Send(priceAndStateRespons); err != nil {
			return err
			log.Printf("Error when streaming back: 'UIPriceAndStateStream'")
			break
		}
		//log.Printf("Sent the following Price and State data: ", priceAndStateRespons)
		time.Sleep(1000 * time.Millisecond)

	}

	log.Println("Leaving stream service!")
	return nil
}

// ******************************************************************************
// Generate allowd gRPC functions
func allowedgRPC_CustomerUI() (gRPCMethodsAllowed *customer_ui_stream_api.RPCMethods) {

	switch customer.CustomerStateMachine.State() {

	case StateCustomerWaitingForCommands:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			true,
			false,
			false,
			false,
			false,
		}

	case StateCustomerPriceHasBeenReceived:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			false,
			true,
			false,
			false,
			false,
		}

	case StateCustomerWaitingForPaymentRequest:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			false,
			false,
			true,
			false,
			false,
		}

	case StateCustomerPaymentRequestReceived:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			false,
			false,
			true,
			false,
			false,
		}

	case StateCustomerHaltedPayments:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			false,
			false,
			false,
			true,
			true,
		}

	default:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			false,
			false,
			false,
			false,
			false,
		}
	}
	return gRPCMethodsAllowed
}
