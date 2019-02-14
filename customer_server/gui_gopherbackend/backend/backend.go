package backend

import (
	"golang.org/x/net/context"
	"jlambert/lightningCab/customer_server/gui_gopherbackend/gui_proto/server"
	"log"
	"time"
	"github.com/sirupsen/logrus"
)

// Backend should be used to implement the server interface
// exposed by the generated server proto.

// Backend should be used to implement the server interface
// exposed by the generated server proto.
//type Backend struct {
//}

type Customer_UI struct {
}

var logger *logrus.Logger


// Ensure struct implements interface
//var _ server.BackendServer = (*Backend)(nil)

var _ server.Customer_UIServer = (*Customer_UI)(nil)

type CallBackFunctionType_AskForPrice func (*server.EmptyParameter) (*server.Price_UI, error)
type CallBackFunctionType_AcceptPrice func (*server.EmptyParameter) (*server.AckNackResponse, error)
type CallBackFunctionType_HaltPayments func (*server.HaltPaymentRequest) (*server.AckNackResponse, error)
type CallBackFunctionType_LeaveTaxi func (*server.EmptyParameter) (*server.AckNackResponse, error)
type CallBackFunctionType_PriceAndStateRespons  func () (*server.UIPriceAndStateRespons, error)

var (
	callbackToCustomer_AskForPrice          CallBackFunctionType_AskForPrice
	callbackToCustomer_AcceptPrice          CallBackFunctionType_AcceptPrice
	callbackToCustomer_HaltPayments         CallBackFunctionType_HaltPayments
	callbackToCustomer_LeaveTaxi            CallBackFunctionType_LeaveTaxi
	callbackToCustomer_PriceAndStateRespons CallBackFunctionType_PriceAndStateRespons

)


// *****************************************************
// Set logger object to same as Customer-logger
func SetLogger(customerLogger *logrus.Logger) {
	logger = customerLogger
}

// *****************************************************
// Set Callback functions

func SetAskForPrice(cbTT CallBackFunctionType_AskForPrice) {
	callbackToCustomer_AskForPrice = cbTT
}
func SetAcceptPrice(cbTT CallBackFunctionType_AcceptPrice) {
	callbackToCustomer_AcceptPrice = cbTT
}
func SetHaltPayments(cbTT CallBackFunctionType_HaltPayments) {
	callbackToCustomer_HaltPayments = cbTT
}
func SetLeaveTaxi(cbTT CallBackFunctionType_LeaveTaxi) {
	callbackToCustomer_LeaveTaxi = cbTT
}

func SetLPriceAndStateRespons(cbTT CallBackFunctionType_PriceAndStateRespons) {
	callbackToCustomer_PriceAndStateRespons = cbTT
}




// ************************************************************************************
// UI-customer Ask for Price

func (s *Customer_UI) AskTaxiForPrice(ctx context.Context, emptyParameter *server.EmptyParameter) (*server.Price_UI, error) {

	log.Println("Incoming from WWW: 'AskTaxiForPrice'")

	returnMessage, err := callbackToCustomer_AskForPrice(emptyParameter)

	return returnMessage, err
}



// ************************************************************************************
// UI-customer accepts price

func (s *Customer_UI) AcceptPrice(ctx context.Context, emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error) {

	log.Println("Incoming from WWW: 'AcceptPrice'")

	returnMessage, err := callbackToCustomer_AcceptPrice(emptyParameter)

	return returnMessage, err

}



// ************************************************************************************
// UI-customer halts payments

func (s *Customer_UI) HaltPayments(ctx context.Context, haltPaymentRequestMessage *server.HaltPaymentRequest) (*server.AckNackResponse, error) {

	log.Println("Incoming from WWW: 'HaltPayments' with parameter: ", haltPaymentRequestMessage.Haltpayment)


	returnMessage, err := callbackToCustomer_HaltPayments(haltPaymentRequestMessage)

	return returnMessage, err

}



// ************************************************************************************
// UI-Customer leaves Taxi

func (s *Customer_UI) LeaveTaxi(ctx context.Context, emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error) {

	log.Println("Incoming: 'LeaveTaxi'")

	returnMessage, err := callbackToCustomer_LeaveTaxi(emptyParameter)

	return returnMessage, err

}

// ************************************************************************************
// Send Stream of Payment-data and state-data for controlling buttons
func (s *Customer_UI) UIPriceAndStateStream(emptyParameter *server.EmptyParameter, stream server.Customer_UI_UIPriceAndStateStreamServer) (err error) {

	log.Printf("Incoming: 'UIPriceAndStateStream'")

	err = nil

	for {

		priceAndStateRespons, err := callbackToCustomer_PriceAndStateRespons()

		if err = stream.Send(priceAndStateRespons); err != nil {
			return err
			log.Printf("Error when streaming back: 'UIPriceAndStateStream'")
			break
		}
		//log.Printf("Sent the following Price and State data: ", priceAndStateRespons)
		time.Sleep(100 * time.Millisecond)




	}

	log.Println("Leaving stream service!")
	return nil
}





