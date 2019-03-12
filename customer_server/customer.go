package main

import (
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"os"
	//taxoTotollGate_api "jlambert/lightningCab/toll_road_server/toll_road_grpc_api"
	"github.com/jlambert68/lightningCab/common_config"
	"github.com/jlambert68/lightningCab/grpc_api/taxi_grpc_api"
	"strconv"
	"time"
	"io"
	"jlambert/lightningCab/customer_server/lightningConnection"
	"github.com/markdaws/simple-state-machine"
	"github.com/btcsuite/btcd/rpcclient"
	"errors"
	"os/signal"
	"syscall"
	//"net"
	//"jlambert/lightningCab/customer_server/customer_html/gopherjs/proto/server"
	//"jlambert/lightningCab/customer_server/customer_html/gopherjs"
	//"github.com/jlambert68/lightningCab/grpc_api/proto/server"
	protoLibrary "jlambert/lightningCab/customer_gui_grpc-web/go/_proto/examplecom/library"
	//"jlambert/lightningCab/vendor_old/github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"jlambert/lightningCab/customer_server/webmain"
)


type Customer struct {
	Title                   string
	CustomerStateMachine    *ssm.StateMachine
	PaymentStreamStarted    bool
	lastRecievedPriceInfo   *taxi_grpc_api.Price
	lastRecievedPriceAccept *taxi_grpc_api.AckNackResponse
	stateBeforeHaltPayments ssm.State
	lastReceivedInvoice     *taxi_grpc_api.PaymentRequest
	receivedTaxiInvoiceButNotPaid bool
	invoiceReceivedAtleastOnce bool
	currentWalletAmountSatoshi int64
	currentWalletAmountSEK float32
	currentTaxiRideAmountSatoshi int64
	currentTaxiRideAmountSEK float32
	averagePaymentAmountSatoshi int64
	averagePaymentAmountSEK float32
	logger *logrus.Logger
	paymentStatistics paymentStatistics_struct
	accelerationPercent int32
	speedPercent int32
}

type averagePayment_struct struct {
	paymentAmount int64
	aggregateAmount int64
	unixTime int64
}
type paymentStatistics_struct struct {
	numbertOfReceivedTransactions int32
	numbertOfPayedTransactions int32
	averageNoOfPayments float32
}

var customer *Customer
var averagePayment []averagePayment_struct

var (
	// Standard Taxi gRPC Server
	remoteTaxiServerConnection *grpc.ClientConn
	customerClient             taxi_grpc_api.TaxiClient

	taxi_address_to_dial string                    = common_config.GrpcTaxiServer_address + common_config.GrpcTaxiServer_port
	useEnv                                         = taxi_grpc_api.TestOrProdEnviroment_Test
	useEnvironment       *taxi_grpc_api.Enviroment = &taxi_grpc_api.Enviroment{TestOrProduction: useEnv}
)

const (
	localCustomerUIRPCServerEngineLocalPort       = common_config.GrpcCustomerUI_RPC_Server_port
	localCustomerUIRPCStreamServerEngineLocalPort = common_config.GrpcCustomerUI_RPC_StreamServer_port
)

/*
var (

//Customer_ui_api
	customerRpcUIServer *grpc.Server
	lis                 net.Listener


	//Customer_ui_stream_api
	customerRpcUIStreamServer *grpc.Server
	lisStream                 net.Listener
)
*/
// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
//type customerUIServiceServer struct{}
type customerUIPriceStreamServiceServer struct{}

func initCustomer() {
	customer = NewCustomer("MyCustomer")
}

func (customer *Customer) startCustomer() {
	var err error
	err = nil
	//customer = NewCustomer("MyCustomer")


	err = validateBitcoind()
	if err != nil {
		////log.Println("Couldn't check Bitcoind, exiting system!")
		customer.logger.Error("Couldn't check Bitcoind, exiting system!")
		os.Exit(0)
	}
	customer.CustomerChecksLightning()
	customer.CustomerIsReadyAndEntersWaitState(false)
}

var (
	// Create States
	StateCustomerInit                     = ssm.State{Name: "StateCustomerInit"}
	StateBitcoinIsCheckedAndOK            = ssm.State{Name: "StateBitcoinIsCheckedAndOK"}
	StateLigtningIsCheckedAndOK           = ssm.State{Name: "StateLigtningIsCheckedAndOK"}
	StateCustomerWaitingForCommands       = ssm.State{Name: "StateCustomerWaitingForCommands"}
	StateCustomerPriceHasBeenReceived     = ssm.State{Name: "StateCustomerPriceHasBeenReceived"}
	StateCustomerWaitingForPaymentRequest = ssm.State{Name: "StateCustomerWaitingForPaymentRequest"}
	StateCustomerPaymentRequestReceived   = ssm.State{Name: "StateCustomerPaymentRequestReceived"}
	StateCustomerHaltedPayments           = ssm.State{Name: "StateCustomerHaltedPayments"}
	StateTaxiKickedOutCustomer            = ssm.State{Name: "StateCustomerKickedOutCustomer"}
	StateCustomerLeftTaxi                 = ssm.State{Name: "StateCustomerLeftTaxi"}
	StateCustomerIsInErrorMode            = ssm.State{Name: "StateCustomerIsInErrorMode"}

	// Create Triggers
	TriggerCustomerChecksBitcoin                    = ssm.Trigger{Key: "TriggerCustomerChecksBitcoin"}
	TriggerCustomerChecksLightning                  = ssm.Trigger{Key: "TriggerCustomerChecksHardware"}
	TriggerCustomerWaitsForCommands                 = ssm.Trigger{Key: "TriggerCustomerWaitsForCommands"}
	TriggerCustomerCommandAskForPrice               = ssm.Trigger{Key: "TriggerCustomerCommandAskForPrice"}
	TriggerCustomerCommandAcceptPrice               = ssm.Trigger{Key: "TriggerCustomerCommandAcceptPrice"}
	TriggerCustomerReceivesPaymentRequest           = ssm.Trigger{Key: "TriggerCustomerReceivesPaymentRequest"}
	TriggerCustomerPaysPaymentRequest               = ssm.Trigger{Key: "TriggerCustomerPaysPaymentRequest"}
	TriggerCustomerCommandHaltPayments              = ssm.Trigger{Key: "TriggerCustomerCommandHaltPayments"}
	TriggerCustomerWillContinueToPay                = ssm.Trigger{Key: "TriggerCustomerWillContinueToPay"}
	TriggerCustomerContiniueToWaitForPaymentRequest = ssm.Trigger{Key: "TriggerCustomerContiniueToWaitForPaymentRequest"}
	TriggerTaxiKicksOutCustomer                     = ssm.Trigger{Key: "TriggerCustomerKicksOutCustomer"}
	TriggerCustomerCommandLeaveTaxi                 = ssm.Trigger{Key: "TriggerCustomerCommandLeaveTaxi"}
	TriggerCustomerEndsInErrorMode                  = ssm.Trigger{Key: "TriggerCustomerEndsInErrorMode"}
)

func NewCustomer(title string) *Customer {

//	currenChannelBalance, err := lightningConnection.CustomerChannelbalance()
//	if err != nil {
//		//log.Println(err)
//	}


	customer := &Customer{
		Title: title,
		PaymentStreamStarted: false,
		invoiceReceivedAtleastOnce:false,
		currentTaxiRideAmountSatoshi: 0,
		currentTaxiRideAmountSEK: 0,
		currentWalletAmountSatoshi: 0, //currenChannelBalance.Balance,
		currentWalletAmountSEK: 0, //SatoshiToSEK(currenChannelBalance.Balance),
		paymentStatistics: paymentStatistics_struct{numbertOfReceivedTransactions:0 , numbertOfPayedTransactions: 0, averageNoOfPayments: 0},
		accelerationPercent: 0,
		speedPercent: 0,
	}

	// Create State machine
	CustomerStateMachine := ssm.NewStateMachine(StateCustomerInit)

	// Configure States: StateCustomerInit
	cfg := CustomerStateMachine.Configure(StateCustomerInit)
	cfg.Permit(TriggerCustomerChecksBitcoin, StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** *** Entering 'StateCustomerInit' ")
	customer.logger.Info("*** *** Entering 'StateCustomerInit' ")
	})
	cfg.OnExit(func() {
		////log.Println("*** Exiting 'StateCustomerInit' ")
		customer.logger.Info("*** Exiting 'StateCustomerInit' ")
		////log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateBitcoinIsCheckedAndOK
	cfg = CustomerStateMachine.Configure(StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerCustomerChecksLightning, StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** *** Entering 'StateBitcoinIsCheckedAndOK' ")
		customer.logger.Info("*** *** Entering 'StateBitcoinIsCheckedAndOK' ")})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		customer.logger.Info("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateLigtningIsCheckedAndOK
	cfg = CustomerStateMachine.Configure(StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateLigtningIsCheckedAndOK' ")
		customer.logger.Info("*** Entering 'StateLigtningIsCheckedAndOK' ")})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		customer.logger.Info("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerWaitingForCommands
	cfg = CustomerStateMachine.Configure(StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerCommandAskForPrice, StateCustomerPriceHasBeenReceived)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerWaitingForCommands' ")
		customer.logger.Info("*** Entering 'StateCustomerWaitingForCommands' ")
		/*if customer.askTaxiForPrice(true) == nil {
			_ = customer.askTaxiForPrice(false)
		}*/
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerWaitingForCommands' ")
		customer.logger.Info("*** Exiting 'StateCustomerWaitingForCommands' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerPriceHasBeenReceived
	cfg = CustomerStateMachine.Configure(StateCustomerPriceHasBeenReceived)
	cfg.Permit(TriggerCustomerCommandAcceptPrice, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerPriceHasBeenReceived' ")
		customer.logger.Info("*** Entering 'StateCustomerPriceHasBeenReceived' ")
		/*if customer.acceptPrice(true) == nil {
			_ = customer.acceptPrice(false)
		}*/
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerPriceHasBeenReceived' ")
		customer.logger.Info("*** Exiting 'StateCustomerPriceHasBeenReceived' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerWaitingForPaymentRequest
	cfg = CustomerStateMachine.Configure(StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerReceivesPaymentRequest, StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerCommandHaltPayments, StateCustomerHaltedPayments)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerWaitingForPaymentRequest' ")
		customer.logger.Info("*** Entering 'StateCustomerWaitingForPaymentRequest' ")
		if customer.PaymentStreamStarted == false {
			go receiveTaxiInvoices(customerClient, useEnvironment)
		}
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerWaitingForPaymentRequest' ")
		customer.logger.Info("*** Exiting 'StateCustomerWaitingForPaymentRequest' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerPaymentRequestReceived
	cfg = CustomerStateMachine.Configure(StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerCommandHaltPayments, StateCustomerHaltedPayments)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerPaymentRequestReceived' ")
		customer.logger.Info("*** Entering 'StateCustomerPaymentRequestReceived' ")

	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerPaymentRequestReceived' ")
		customer.logger.Info("*** Exiting 'StateCustomerPaymentRequestReceived' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerHaltedPayments
	cfg = CustomerStateMachine.Configure(StateCustomerHaltedPayments)
	cfg.Permit(TriggerCustomerWillContinueToPay, StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerCommandLeaveTaxi, StateCustomerLeftTaxi)
	cfg.Permit(TriggerCustomerContiniueToWaitForPaymentRequest, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerHaltedPayments' ")
		customer.logger.Info("*** Entering 'StateCustomerHaltedPayments' ")
		//_ = customer.CustomerStateMachine.Fire(TriggerCustomerSetsHardwareInDriveMode.Key, nil)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerHaltedPayments' ")
		customer.logger.Info("*** Exiting 'StateCustomerHaltedPayments' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateTaxiKickedOutCustomer
	cfg = CustomerStateMachine.Configure(StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateTaxiKickedOutCustomer' ")
		customer.logger.Info("*** Entering 'StateTaxiKickedOutCustomer' ")})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxiKickedOutCustomer' ")
		customer.logger.Info("*** Exiting 'StateTaxiKickedOutCustomer' ")
		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerLeftTaxi
	cfg = CustomerStateMachine.Configure(StateCustomerLeftTaxi)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateCustomerLeftTaxi' ")
		customer.logger.Info("*** Entering 'StateCustomerLeftTaxi' ")})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerLeftTaxi' ")
		customer.logger.Info("*** Exiting 'StateCustomerLeftTaxi' ")

		//log.Println("")
		customer.logger.Info("")
	})

	// Configure States: StateCustomerIsInErrorMode
	cfg = CustomerStateMachine.Configure(StateCustomerIsInErrorMode)

	cfg.OnEnter(func() { //log.Println("*** Entering 'StateCustomerIsInErrorMode' ")
		customer.logger.Info("*** Entering 'StateCustomerIsInErrorMode' ")})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerIsInErrorMode' ")
		customer.logger.Info("*** Exiting 'StateCustomerIsInErrorMode' ")
		//log.Println("")
		customer.logger.Info("")
	})

	customer.CustomerStateMachine = CustomerStateMachine

	return customer
}

// ******************************************************************************
// log Errors for Triggers and States
func logTriggerStateError(spaceCount int, currentState ssm.State, trigger ssm.Trigger, err error) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, "Current state:", currentState, " doesn't accept trigger'", trigger, "'. Error Message: ", err)
	messageToLog := "Current state: " + currentState.Name + " doesn't accept trigger'" + trigger.String() + "'."
	customer.logger.Info(messageToLog)
	}

// ******************************************************************************

// ******************************************************************************
// Log Errors for Triggers and States
 func logMessagesWithOutError(spaceCount int, message string) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, message)
	customer.logger.Info(message)
}

// ******************************************************************************
// Log Errors for Triggers and States
func logMessagesWithError(spaceCount int, message string, err error) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, message, err)
	customer.logger.WithFields(logrus.Fields{
		"error": err,
	}).Info(message)
}

// ******************************************************************************
// Ask Taxi for Price
func CallBackAskTaxiForPrice(emptyParameter *protoLibrary.EmptyParameter) (*protoLibrary.Price_UI, error) {
	//log.Println("Incoming: 'AskTaxiForPrice'")
	customer.logger.Info("Incoming: 'AskTaxiForPrice'")

	returnMessage := &protoLibrary.Price_UI{
		Acknack:                   false,
		Comments:                  "",
		SpeedAmountSatoshi:        0,
		AccelerationAmountSatoshi: 0,
		TimeAmountSatoshi:         0,
		SpeedAmountSek:            0,
		AccelerationAmountSek:     0,
		TimeAmountSek:             0,
		Timeunit:                  0,
		PaymentRequestInterval:    0,
		Priceunit:                 0,
	}

	// Check if State machine accepts State change
	err := customer.askTaxiForPrice(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.askTaxiForPrice(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &protoLibrary.Price_UI{
					Acknack:                   true,
					Comments:                  customer.lastRecievedPriceInfo.Comments,
					SpeedAmountSatoshi:        customer.lastRecievedPriceInfo.GetSpeed(),
					AccelerationAmountSatoshi: customer.lastRecievedPriceInfo.GetAcceleration(),
					TimeAmountSatoshi:         customer.lastRecievedPriceInfo.GetTime(),
					SpeedAmountSek:            float32(customer.lastRecievedPriceInfo.GetSpeed()) * common_config.BTCSEK,
					AccelerationAmountSek:     float32(customer.lastRecievedPriceInfo.GetAcceleration()) * common_config.BTCSEK,
					TimeAmountSek:             float32(customer.lastRecievedPriceInfo.GetTime()) * common_config.BTCSEK,
					Timeunit:                  0, //customer.lastRecievedPriceInfo.Timeunit,
					PaymentRequestInterval:    0, //customer.lastRecievedPriceInfo.PaymentRequestInterval,
					Priceunit:                 0, //customer.lastRecievedPriceInfo.Priceunit,
				}

			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"
	}
	return returnMessage, nil
}


// Code for State Change for askTaxiForPrice
func (customer *Customer) askTaxiForPrice(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerCommandAskForPrice

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger

		resp, err := customerClient.AskTaxiForPrice(context.Background(), useEnvironment)
		if err != nil {
			logMessagesWithError(4, "Could not send 'AskTaxiForPrice' to address: "+taxi_address_to_dial+". Error Message:", err)
			break
		} else {

			//Save last Price respons
			customer.lastRecievedPriceInfo = resp

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'AskTaxiForPrice' on address "+taxi_address_to_dial+" successfully processed")
				logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)
				logMessagesWithOutError(4, "Response Message (Time): "+strconv.Itoa(int(resp.Time)))
				logMessagesWithOutError(4, "Response Message (Speed): "+strconv.Itoa(int(resp.Speed)))
				logMessagesWithOutError(4, "Response Message (Acceleration): "+strconv.Itoa(int(resp.Acceleration)))
				logMessagesWithOutError(4, "Response Message (PaymentRequestInterval): "+strconv.Itoa(int(resp.PaymentRequestInterval)))
				logMessagesWithOutError(4, "Response Message (Priceunit): "+strconv.Itoa(int(resp.Priceunit)))
				logMessagesWithOutError(4, "Response Message (Timeunit): "+strconv.Itoa(int(resp.Timeunit)))

			} else {
				logMessagesWithOutError(4, "'AskTaxiForPrice' on address "+taxi_address_to_dial+" NOT successfully processed")
				logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)
				logMessagesWithOutError(4, "Response Message (Time): "+strconv.Itoa(int(resp.Time)))
				logMessagesWithOutError(4, "Response Message (Speed): "+strconv.Itoa(int(resp.Speed)))
				logMessagesWithOutError(4, "Response Message (Acceleration): "+strconv.Itoa(int(resp.Acceleration)))
				logMessagesWithOutError(4, "Response Message (PaymentRequestInterval): "+strconv.Itoa(int(resp.PaymentRequestInterval)))
				logMessagesWithOutError(4, "Response Message (Priceunit): "+strconv.Itoa(int(resp.Priceunit)))
				logMessagesWithOutError(4, "Response Message (Timeunit): "+strconv.Itoa(int(resp.Timeunit)))

				err = errors.New("'AskTaxiForPrice' on address " + taxi_address_to_dial + " NOT successfully processed")
			}
		}

		if err == nil && resp.GetAcknack() == true {
			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
			if err != nil {
				logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

			}
		}
	}

	return err

}


// ******************************************************************************
// Customer accepts Price

func CallBackAcceptPrice(emptyParameter *protoLibrary.EmptyParameter) (*protoLibrary.AckNackResponse, error){

	//log.Println("Incoming: 'AcceptPrice'")
	customer.logger.Info("Incoming: 'AcceptPrice'")

	returnMessage := &protoLibrary.AckNackResponse{
		Acknack:  false,
		Comments: "",
	}

	// Check if State machine accepts State change
	err := customer.acceptPrice(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.acceptPrice(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &protoLibrary.AckNackResponse{
					Acknack:  true,
					Comments: customer.lastRecievedPriceAccept.Comments,
				}
			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer accept price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer accept price"
	}
	return returnMessage, nil
}


// Code for State Change for acceptPrice
func (customer *Customer) acceptPrice(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerCommandAcceptPrice

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger

		resp, err := customerClient.AcceptPrice(context.Background(), useEnvironment)
		if err != nil {
			logMessagesWithError(4, "Could not send 'AcceptPrice' to address: "+taxi_address_to_dial+". Error Message:", err)
			break

		} else {

			//Save last PriceAccept respons
			customer.lastRecievedPriceAccept = resp

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'AcceptPrice' on address "+taxi_address_to_dial+" successfully processed")
				logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)

				// Moved to state machine ---go receiveTaxiInvoices(customerClient, useEnvironment)

			} else {
				logMessagesWithOutError(4, "'AcceptPrice' on address "+taxi_address_to_dial+" NOT successfully processed")
				logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)

				err = errors.New("'AcceptPrice' on address " + taxi_address_to_dial + " NOT successfully processed")
			}
		}

		if err == nil && resp.GetAcknack() == true {
			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
			if err != nil {
				logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

			}
		}
	}

	return err

}

// ******************************************************************************
// Customer Halts Payments
func CallBackHaltPayments(haltPaymentRequestMessage *protoLibrary.HaltPaymentRequest) (*protoLibrary.AckNackResponse, error){

	//log.Println("Incoming: 'HaltPayments' with parameter: ", haltPaymentRequestMessage.Haltpayment)
	customer.logger.WithFields(logrus.Fields{
		"haltPaymentRequestMessage.Haltpaymen":    haltPaymentRequestMessage.Haltpayment,
	}).Info("Incoming: 'HaltPayments' with parameter!")

	returnMessage := &protoLibrary.AckNackResponse{
		Acknack:  false,
		Comments: "",
	}

	// Decide of user wants to Halt or un-Halt payment
	switch haltPaymentRequestMessage.Haltpayment {

	case true: //Halt Payments
		// Check if State machine accepts State change
		err := customer.haltPayments(true)

		if err == nil {

			switch  err.(type) {
			case nil:
				err = customer.haltPayments(false)
				if err != nil {
					logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
					returnMessage.Comments = "State machine is not in correct state to be able to halt payment"

				} else {

					logMessagesWithOutError(4, "Success in change if state: ")

					returnMessage = &protoLibrary.AckNackResponse{
						Acknack:  true,
						Comments: "Success in change of state and Halt payments",
					}
				}

			default:
				logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able to halt payment"

			}
		} else {
			logMessagesWithError(4, "State machine is not in correct state to be able to halt payment: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able to halt payment"
		}

	case false: //Unhalt payments
		if customer.CustomerStateMachine.IsInState(StateCustomerHaltedPayments) == true {

			// Correct state for Un-Halt payments
			switch customer.stateBeforeHaltPayments {

			// Go back to wait for PaymentRequest
			case StateCustomerWaitingForPaymentRequest:
				currentTrigger := TriggerCustomerContiniueToWaitForPaymentRequest

				err := customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
				if err != nil {
					logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)
					returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
				} else {
					returnMessage = &protoLibrary.AckNackResponse{
						Acknack:  true,
						Comments: "Success in change of state and Un-Halt payments",
					}
				}

				// Go back and pay paymentRequest
			case StateCustomerPaymentRequestReceived:
				currentTrigger := TriggerCustomerWillContinueToPay

				err := customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
				if err != nil {
					logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)
					returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
				} else {
					returnMessage = &protoLibrary.AckNackResponse{
						Acknack:  true,
						Comments: "Success in change of state and Un-Halt payments",
					}
				}
			}

		} else {

			// Not Correct state to be able of un-halting payments
			logMessagesWithOutError(4, "State machine is not in correct state to be able to un-halt payment")
			returnMessage.Comments = "State machine is not in correct state to be able to un-halt payment"
		}

	}
	return returnMessage, nil
}

// Function for change state
func (customer *Customer) haltPayments(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerCommandHaltPayments

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		customer.stateBeforeHaltPayments = customer.CustomerStateMachine.State()
		// Execute Trigger
		err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Customer Leaves Taxi

func CallBackLeaveTaxi(emptyParameter *protoLibrary.EmptyParameter) (*protoLibrary.AckNackResponse, error){

	//log.Println("Incoming: 'LeaveTaxi'")
	customer.logger.Info("Incoming: 'LeaveTaxi'")

	returnMessage := &protoLibrary.AckNackResponse{
		Acknack:  false,
		Comments: "",
	}

	// Check if State machine accepts State change
	err := customer.leaveTaxi(true)

	if err == nil {

		switch  err.(type) {
		case nil:
			err = customer.leaveTaxi(false)
			if err != nil {
				logMessagesWithError(4, "State machine is not in correct state to be able have customer leave taxi: ", err)
				returnMessage.Comments = "State machine is not in correct state to be able have customer leave taxi"

			} else {

				logMessagesWithOutError(4, "Success in change if state: ")

				returnMessage = &protoLibrary.AckNackResponse{
					Acknack:  true,
					Comments: "State machine is in correct state and customer left taxi",
				}
			}

		default:
			logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
			returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"

		}
	} else {
		logMessagesWithError(4, "State machine is not in correct state to be able have customer ask for price: ", err)
		returnMessage.Comments = "State machine is not in correct state to be able have customer ask for price"
	}
	return returnMessage, nil
}

// Functione for state change
func (customer *Customer) leaveTaxi(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerCommandLeaveTaxi

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		customer.stateBeforeHaltPayments = customer.CustomerStateMachine.State()
		// Execute Trigger
		err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}


// ******************************************************************************
// Customer recieves paymentRequest Stream
func receiveTaxiInvoices(client taxi_grpc_api.TaxiClient, enviroment *taxi_grpc_api.Enviroment) {
	//log.Printf("Starting Taxi PaymentRequest stream %v", enviroment)
	customer.logger.Info("Starting Taxi PaymentRequest stream")

	var invoicesAndPaymentData []*taxi_grpc_api.PaymentRequest

	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	//defer cancel()

	stream, err := client.PaymentRequestStream(ctx, enviroment)
	if err != nil {
		//log.Fatalf("Problem to connect to Taxi Invoice Stream: ", client, err)
		customer.logger.WithFields(logrus.Fields{
			"client":    client,
			"err": err,
		}).Error("Problem to connect to Taxi Invoice Stream")
		os.Exit(0)
	}
	// Used for controlling state machine
	customer.PaymentStreamStarted = true
	for {
		if customer.CustomerStateMachine.IsInState(StateCustomerWaitingForPaymentRequest) {

			invoiceAndPaymentData, err := stream.Recv()
			if err == io.EOF {
				//log.Println("HMMM, skumt borde inte slutat här när vi tar emot InvoiceStream, borde avsluta Customer")
				customer.logger.Error("MMM, skumt borde inte slutat här när vi tar emot InvoiceStream, borde avsluta Customer")
				customer.PaymentStreamStarted = false
				break
			}
			if err != nil {
				//log.Fatalf("Problem when streaming from Taxi invoiceAndPaymentData Stream:", client, err)
				customer.logger.WithFields(logrus.Fields{
					"client":    client,
					"err": err,
				}).Error("Problem when streaming from Taxi invoiceAndPaymentData Stream")
				os.Exit(0)

			}
			customer.paymentStatistics.numbertOfReceivedTransactions = customer.paymentStatistics.numbertOfReceivedTransactions + 1
			customer.receivedTaxiInvoiceButNotPaid = true
			customer.invoiceReceivedAtleastOnce = true

			customer.accelerationPercent = invoiceAndPaymentData.AccelerationPercent
			customer.speedPercent = invoiceAndPaymentData.SpeedPercent

			//Customer Pays Invoice
			invoicesAndPaymentData = append(invoicesAndPaymentData, invoiceAndPaymentData)

			customer.lastReceivedInvoice = invoiceAndPaymentData
			customer.currentTaxiRideAmountSatoshi = customer.currentTaxiRideAmountSatoshi + invoiceAndPaymentData.TotalAmountSatoshi
			customer.currentTaxiRideAmountSEK = customer.currentTaxiRideAmountSEK + invoiceAndPaymentData.TotalAmountSek
			/*
						customer.lastReceivedInvoice.AccelerationAmountSatoshi = invoiceAndPaymentData.AccelerationAmountSatoshi
						customer.lastReceivedInvoice.SpeedAmountSatoshi = invoiceAndPaymentData.SpeedAmountSatoshi
						customer.lastReceivedInvoice.TimeAmountSatoshi = invoiceAndPaymentData.TimeAmountSatoshi
						customer.lastReceivedInvoice.AccelerationAmountSek = invoiceAndPaymentData.AccelerationAmountSek
						customer.lastReceivedInvoice.SpeedAmountSek = invoiceAndPaymentData.SpeedAmountSek
						customer.lastReceivedInvoice.TimeAmountSek = invoiceAndPaymentData.TimeAmountSek
						customer.lastReceivedInvoice.TotalAmountSatoshi = customer.lastReceivedInvoice.TotalAmountSatoshi + invoiceAndPaymentData.TotalAmountSatoshi
						customer.lastReceivedInvoice.TotalAmountSek = customer.lastReceivedInvoice.TotalAmountSek + invoiceAndPaymentData.TotalAmountSek
			*/

			//log.Println("Invoices: ", invoicesAndPaymentData)
			customer.logger.WithFields(logrus.Fields{
				"Invoices":    invoicesAndPaymentData,
			}).Info("Invoices")

			stateChangeResponse := customer.CustomerStateMachine.Fire(TriggerCustomerReceivesPaymentRequest.Key, nil)
			if stateChangeResponse != nil {
				//log.Println("Error, Should be able to change state with Trigger: 'TriggerCustomerReceivesPaymentRequest'")
				customer.logger.Warning("Error, Should be able to change state with Trigger: 'TriggerCustomerReceivesPaymentRequest'")
			} else {

				if customer.CustomerStateMachine.IsInState(StateCustomerPaymentRequestReceived) {
					err = lightningConnection.PayReceivedInvoicesFromTaxi(invoicesAndPaymentData)
					if err != nil {
						//log.Println("Problem when paying Invoice from Taxi: ", err)
						customer.logger.WithFields(logrus.Fields{
							"err":    err,
						}).Warning("Problem when paying Invoice from Taxi")
						customer.PaymentStreamStarted = false
						break
					} else {
						// Update info about Wallet Channel Balance
						walletRespons, err := lightningConnection.CustomerChannelbalance()
						if err != nil {
							//log.Println("Error, couldn't get Wallet Balance")
							customer.logger.Warning("Error, couldn't get Wallet Balance")
						} else {
							customer.currentWalletAmountSatoshi = walletRespons.Balance
							customer.currentWalletAmountSEK = SatoshiToSEK(customer.currentWalletAmountSatoshi)
						}

						// Add payment amount to average calculation
						customer.averagePaymentAmountSatoshi, customer.paymentStatistics.averageNoOfPayments = calculateAveragePaymentAmount(invoiceAndPaymentData.TotalAmountSatoshi)
						customer.averagePaymentAmountSEK = SatoshiToSEK(customer.averagePaymentAmountSatoshi)

						//log.Println("Invoice from Taxi-Stream is paid: ", invoicesAndPaymentData)
						customer.logger.WithFields(logrus.Fields{
							"invoicesAndPaymentData":    invoicesAndPaymentData,
						}).Info("Invoice from Taxi-Stream is paid")
						customer.paymentStatistics.numbertOfPayedTransactions = customer.paymentStatistics.numbertOfPayedTransactions + 1

						// Remove paid invoice
						invoicesAndPaymentData = append(invoicesAndPaymentData[:0], invoicesAndPaymentData[1:]...)

						customer.receivedTaxiInvoiceButNotPaid = true

					}

					stateChangeResponse = customer.CustomerStateMachine.Fire(TriggerCustomerPaysPaymentRequest.Key, nil)
					if stateChangeResponse != nil {
						//log.Println("Error, Should be able to change state with Trigger: 'TriggerCustomerPaysPaymentRequest'")
						customer.logger.Warning("Error, Should be able to change state with Trigger: 'TriggerCustomerPaysPaymentRequest'")
					}
				}
			}
		} else {

			customer.PaymentStreamStarted = false
			break
		}
	}

	customer.PaymentStreamStarted = false
}

func SatoshiToSEK( satoshis int64) (sek float32) {
	sek = float32(satoshis) * common_config.BTCSEK / common_config.SatoshisPerBTC
	return sek
}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (customer *Customer) CustomerCheckBitcoin(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerChecksBitcoin

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Check(or go to) if Cusomter can go to wait state
func (customer *Customer) CustomerIsReadyAndEntersWaitState(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerWaitsForCommands

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if customer.CustomerStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (customer *Customer) CustomerChecksLightning() {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerChecksLightning

	err := customer.CustomerStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, customer.CustomerStateMachine.State(), currentTrigger, err)
	}

}

// *****************************************************************************

// ******************************************************************************

func validateBitcoind() (err error) {

	// Check if State machine accepts State change
	err = customer.CustomerCheckBitcoin(true)

	// If OK to change State then check Bitcoind
	if err == nil {

		// create new client instance
		client, err := rpcclient.New(&rpcclient.ConnConfig{
			HTTPPostMode: true,
			DisableTLS:   true,
			Host:         "127.0.0.1:18332",
			User:         "jlambert",
			Pass:         "jlambert97531",
		}, nil)
		if err != nil {
			//log.Fatalf("error creating new btc client: %v", err)
			customer.logger.WithFields(logrus.Fields{
				"err":    err,
			}).Error("Error creating new btc client")
			os.Exit(0)

		}

		// If SimNet is used then skip this check
		if common_config.UseSimnet == false {
			// Wait for Blockchain to Sync
			for {

				// Get Blockchain info and verifiction status
				blockchainInfo, _ := client.GetBlockChainInfo()
				headers := blockchainInfo.Headers
				blocks := blockchainInfo.Blocks

				// If not 100% sync then wait
				if headers != blocks {
					logMessagesWithOutError(4, "Bitcoin blockchain not synced, waiting for 30 seconds...")
					logMessagesWithOutError(4, "Current Block"+strconv.Itoa(int(blocks))+", Headers: "+strconv.Itoa(int(headers)))

					time.Sleep(30 * time.Second)

				} else {

					logMessagesWithOutError(4, "Bitcoin blockchain is fully synced")
					blockheight := blockchainInfo.Blocks
					logMessagesWithOutError(4, "Blockheight: "+strconv.Itoa(int(blockheight)))
					break

				}
			}
		}
	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)

	}

	err = customer.CustomerCheckBitcoin(false)
	if err != nil {
		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)
	}

	return err
}

// ******************************************************************************
// Calculate Average Payment
func calculateAveragePaymentAmount(latestAmount int64) (int64, float32) {
	var latestPaymentAmount averagePayment_struct

	// Get current Unix time in seconds
	now := time.Now()
	secs := now.Unix()

	latestPaymentAmount.paymentAmount = latestAmount
	latestPaymentAmount.unixTime = secs

	if len(averagePayment) != 0 {
		for {
			// Remove all data objects that has an old timestamp
			if secs-averagePayment[0].unixTime > common_config.TimeForAveragePaymentCalculation {
				// Subract old data from latest and Remove post due to old data
				averagePayment[len(averagePayment)-1].aggregateAmount = averagePayment[len(averagePayment)-1].aggregateAmount - averagePayment[0].paymentAmount
				averagePayment = averagePayment[1:]
			} else {

				break
			}
		}
	}

		// Add the new payment object
		if len(averagePayment) == 0 {
			// If empty then this is the first object
			latestPaymentAmount.aggregateAmount = latestAmount
			averagePayment = append(averagePayment, latestPaymentAmount)

		} else {
			latestPaymentAmount.aggregateAmount = averagePayment[len(averagePayment)-1].aggregateAmount + latestAmount
			averagePayment = append(averagePayment, latestPaymentAmount)
		}

	averagePaymentValue := averagePayment[len(averagePayment)-1].aggregateAmount / int64(len(averagePayment))
	averageNoOfPayments := float32(len(averagePayment)) / common_config.TimeForAveragePaymentCalculation
	return averagePaymentValue, averageNoOfPayments
}

// ******************************************************************************
// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		//log.Println("Clean up and shut down servers")
		customer.logger.Info("Clean up and shut down servers")

		remoteTaxiServerConnection.Close()

		/*
		//log.Println("Gracefull stop for: 'customerRpcUIServer'")
		customerRpcUIprotoLibrary.GracefulStop()
*/
//		//log.Println("Gracefull stop for: 'customerRpcUIStreamServer'")
//		customerRpcUIStreamprotoLibrary.GracefulStop()
/*
		//log.Println("Close net.Listing: %v", localCustomerUIRPCServerEngineLocalPort)
		lis.Close()
*/
//		//log.Println("Close net.Listing: %v", localCustomerUIRPCStreamServerEngineLocalPort)
//		lisStream.Close()

	}
}

func main() {
/*
	agent := stackimpact.Start(stackimpact.Options{
		AgentKey: "3595d3d1600b648248cebaf5bc0f9be8c6b4a74e",
		AppName: "LightningCAB - Customer",
	})

	span := agent.Profile();
	defer span.Stop();
*/
	var err error

	defer cleanup()

	// Init Customer Object
	initCustomer()

	// Init logger
	customer.InitLogger("")

	// Should only be done from init functions
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(customer.logger.Out, customer.logger.Out, customer.logger.Out))
	// *********************
	// Set up connection to Toll Gate Hardware Server
	remoteTaxiServerConnection, err = grpc.Dial(taxi_address_to_dial, grpc.WithInsecure())
	if err != nil {
		customer.logger.WithFields(logrus.Fields{
			"taxi_address_to_dial":    taxi_address_to_dial,
			"error message": err,
		}).Error("Did not connect to Taxi Server!")
		// //log.Println("did not connect to Taxi Server on address: ", taxi_address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		customer.logger.WithFields(logrus.Fields{
			"taxi_address_to_dial":    taxi_address_to_dial,
		}).Info("gRPC connection OK to Taxi Server!")
		////log.Println("gRPC connection OK to Taxi Server, address: ", taxi_address_to_dial)
		// Creates a new Clients
		customerClient = taxi_grpc_api.NewTaxiClient(remoteTaxiServerConnection)

	}

	//Init Lightning
	go lightningConnection.LigtningMainService(customer.logger)
	time.Sleep(5 * time.Second)

	r, e := lightningConnection.CustomerWalletbalance()
	//spew.Println(r, e)
	customer.logger.WithFields(logrus.Fields{
		"Wallet Balance":    r,
		"err": e,
	}).Info("lightningConnection.CustomerWalletbalance()")
	r2, e2 := lightningConnection.CustomerChannelbalance()
	//spew.Println(r2, e2)
	customer.logger.WithFields(logrus.Fields{
		"Channel Balance":    r2,
		"err": e2,
	}).Info("lightningConnection.CustomerChannelbalance()")

	//Start Web Backend
	go webmain.Webmain(customer.logger,
		CallBackAskTaxiForPrice,
		CallBackAcceptPrice,
		CallBackHaltPayments,
		CallBackLeaveTaxi,
		CallBackPriceAndStateData)
/*
	// *********************
	// Start Customer RPC UI Server for Incomming web connectionss
	//log.Println("Customer UI RPC Server started")
	//log.Println("Start listening on: %v", localCustomerUIRPCServerEngineLocalPort)
	lis, err = net.Listen("tcp", localCustomerUIRPCServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
*/
	//*****************

/*
	// Start Customer RPC UI Stream Server for Incomming web connectionss
	//log.Println("Customer UI RPC Stream Server started")
	//log.Println("Start listening on: %v", localCustomerUIRPCStreamServerEngineLocalPort)
	lisStream, err = net.Listen("tcp", localCustomerUIRPCStreamServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
*/

/*
	// Creates a new RegisterClient gRPC server for UI
	go func() {
		//log.Println("Starting Customer RPC UI Server")
		customerRpcUIServer = grpc.NewServer()
		//customerRpcUIServer = grpc.NewServer(grpc.Creds(creds))
		customer_ui_api.RegisterCustomer_UIServer(customerRpcUIServer, &customerUIServiceServer{})
		//log.Println("registerTaxiServer for Taxi Gate started")
		//log.Println(customerRpcUIprotoLibrary.GetServiceInfo())
		customerRpcUIprotoLibrary.Serve(lis)


	}()
*/

/*

	// Creates a new RegisterClient gRPC stream server for UI
	go func() {
		//log.Println("Starting Customer RPC UI Stream Server")
		customerRpcUIStreamServer = grpc.NewServer()
		customer_ui_stream_api.RegisterCustomerUIPriceStreamServer(customerRpcUIStreamServer, &customerUIPriceStreamServiceServer{})
		//log.Println("registerTaxiServer for Taxi Gate started")
		customerRpcUIStreamprotoLibrary.Serve(lisStream)
	}()
	// *********************

*/

	// Set up the Private Taxi Road State Machine
	customer.startCustomer()


	//*****************

	// Set system in wait mode for externa input ...
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	for {
		customer.logger.Info("Sleeping...for another 5 minutes!")
		//fmt.Println("sleeping...for another 5 minutes")
		time.Sleep(300 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}

}
