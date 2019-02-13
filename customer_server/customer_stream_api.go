package main

import (

	"github.com/jlambert68/lightningCab/grpc_api/proto/server"
	"time"
)

func CallBackPriceAndStateData() (*server.UIPriceAndStateRespons, error) {
	var err error
	err = nil
	var priceAndStateRespons *server.UIPriceAndStateRespons

	//Check if any InvoinceData has been received
	if customer.invoiceReceivedAtleastOnce == false {
		priceAndStateRespons = &server.UIPriceAndStateRespons{
			Acknack:                   true,
			Comments:                  "",
			SpeedAmountSatoshi:        0,
			AccelerationAmountSatoshi: 0,
			TimeAmountSatoshi:         0,
			SpeedAmountSek:            0,
			AccelerationAmountSek:     0,
			TimeAmountSek:             0,
			TotalAmountSatoshi:        0,
			TotalAmountSek:            0,
			Timestamp:                 time.Now().UnixNano(),
			AllowedRPCMethods:         allowedgRPC_CustomerUI(),
			CurrentTaxirideSatoshi:	0,
			CurrentTaxiRideSek: 0,
			CurrentWalletbalanceSatoshi: int64(customer.currentWalletAmountSatoshi),
			CurrentWalletbalanceSek: float32(customer.currentWalletAmountSEK),
			AveragePaymentAmountSatoshi: 0,
			AveragePaymentAmountSek: 0,
			AvaregeNumberOfPayments: 0,
			AccelarationPercent:0,
			SpeedPercent:0,
		}
	} else {
		priceAndStateRespons = &server.UIPriceAndStateRespons{
			Acknack:                   true,
			Comments:                  "",
			SpeedAmountSatoshi:        int64(customer.lastReceivedInvoice.SpeedAmountSatoshi),
			AccelerationAmountSatoshi: int64(customer.lastReceivedInvoice.AccelerationAmountSatoshi),
			TimeAmountSatoshi:         int64(customer.lastReceivedInvoice.TimeAmountSatoshi),
			SpeedAmountSek:            float32(customer.lastReceivedInvoice.SpeedAmountSek),
			AccelerationAmountSek:     float32(customer.lastReceivedInvoice.AccelerationAmountSek),
			TimeAmountSek:             float32(customer.lastReceivedInvoice.TimeAmountSek),
			TotalAmountSatoshi:        int64(customer.lastReceivedInvoice.TotalAmountSatoshi),
			TotalAmountSek:            float32(customer.lastReceivedInvoice.TotalAmountSek),
			Timestamp:                 time.Now().UnixNano(),
			AllowedRPCMethods:         allowedgRPC_CustomerUI(),
			CurrentTaxirideSatoshi:	int64(customer.currentTaxiRideAmountSatoshi),
			CurrentTaxiRideSek: float32(customer.currentTaxiRideAmountSEK),
			CurrentWalletbalanceSatoshi: int64(customer.currentWalletAmountSatoshi),
			CurrentWalletbalanceSek: float32(customer.currentWalletAmountSEK),
			AveragePaymentAmountSatoshi: int64(customer.averagePaymentAmountSatoshi),
			AveragePaymentAmountSek: float32(customer.averagePaymentAmountSEK),
			AvaregeNumberOfPayments: float32(customer.paymentStatistics.averageNoOfPayments),
			AccelarationPercent: customer.accelerationPercent,
			SpeedPercent: customer.speedPercent,
		}
	}

	return priceAndStateRespons, err

}

/*
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
*/

// ******************************************************************************
// Generate allowd gRPC functions
func allowedgRPC_CustomerUI() (gRPCMethodsAllowed *server.RPCMethods) {

	switch customer.CustomerStateMachine.State() {

	case StateCustomerWaitingForCommands:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   true,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerPriceHasBeenReceived:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       true,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerWaitingForPaymentRequest:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  true,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerPaymentRequestReceived:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  true,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}

	case StateCustomerHaltedPayments:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: true,
			LeaveTaxi:         true,
		}

	default:
		gRPCMethodsAllowed = &server.RPCMethods{
			AskTaxiForPrice:   false,
			AcceptPrice:       false,
			HaltPaymentsTrue:  false,
			HaltPaymentsFalse: false,
			LeaveTaxi:         false,
		}
	}
	return gRPCMethodsAllowed
}
