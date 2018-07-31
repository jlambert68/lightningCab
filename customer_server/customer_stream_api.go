package main

import (
	"log"
	"time"
	"jlambert/lightningCab/customer_server/customer_ui_stream_api"
)

func (s *customerUIPriceStreamServiceServer) UIPriceAndStateStream(emptyParameter *customer_ui_stream_api.EmptyParameter, stream customer_ui_stream_api.CustomerUIPriceStream_UIPriceAndStateStreamServer) (err error) {
	log.Printf("Incoming: 'UIPriceAndStateStream'")

	err = nil

	for {
		priceAndStateRespons := &customer_ui_stream_api.UIPriceAndStateRespons{
			Acknack:                   true,
			Comments:                  "",
			SpeedAmountSatoshi:        customer.lastReceivedInvoice.SpeedAmountSatoshi,
			AccelerationAmountSatoshi: customer.lastReceivedInvoice.AccelerationAmountSatoshi,
			TimeAmountSatoshi:         customer.lastReceivedInvoice.TimeAmountSatoshi,
			SpeedAmountSek:            customer.lastReceivedInvoice.SpeedAmountSek,
			AccelerationAmountSek:     customer.lastReceivedInvoice.AccelerationAmountSek,
			TimeAmountSek:             customer.lastReceivedInvoice.TimeAmountSek,
			TotalAmountSatoshi:        customer.lastReceivedInvoice.TotalAmountSatoshi,
			TotalAmountSek:            customer.lastReceivedInvoice.TotalAmountSek,
			Timestamp:                 time.Now().UnixNano(),
			AllowedRPCMethods:         allowedgRPC_CustomerUI(),
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
			AskTaxiForPrice:   true,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerPriceHasBeenReceived:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       true,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerWaitingForPaymentRequest:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  true,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerPaymentRequestReceived:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  true,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerHaltedPayments:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: true,
			LeaveTaxi:         true,
		}

	default:
		gRPCMethodsAllowed = &customer_ui_stream_api.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}
	}
	return gRPCMethodsAllowed
}
