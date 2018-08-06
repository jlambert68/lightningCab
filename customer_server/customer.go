package main

import (
	"strings"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"log"
	"os"
	//taxoTotollGate_api "jlambert/lightningCab/toll_road_server/toll_road_grpc_api"
	"jlambert/lightningCab/common_config"
	"jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"strconv"
	"time"
	"io"
	"jlambert/lightningCab/customer_server/lightningConnection"
	"github.com/markdaws/simple-state-machine"
	"github.com/btcsuite/btcd/rpcclient"
	"errors"
	"os/signal"
	"syscall"
	"fmt"
	"net"
	"jlambert/lightningCab/customer_server/customer_ui_stream_api"
	"jlambert/lightningCab/customer_server/customer_html/gopherjs/proto/server"
	"jlambert/lightningCab/customer_server/customer_html/gopherjs"
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
}

var customer *Customer

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

var (
	/*
	//Customer_ui_api
	customerRpcUIServer *grpc.Server
	lis                 net.Listener
*/

	//Customer_ui_stream_api
	customerRpcUIStreamServer *grpc.Server
	lisStream                 net.Listener
)

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
//type customerUIServiceServer struct{}
type customerUIPriceStreamServiceServer struct{}

func initiateCustomer() {
	var err error
	err = nil
	customer = NewCustomer("MyCustomer")
	//customer.RestartToaxiSystem()

	err = validateBitcoind()
	if err != nil {
		log.Println("Couldn't check Bitcoind, exiting system!")
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

	customer := &Customer{Title: title, PaymentStreamStarted: false}
	// Create State machine
	CustomerStateMachine := ssm.NewStateMachine(StateCustomerInit)

	// Configure States: StateCustomerInit
	cfg := CustomerStateMachine.Configure(StateCustomerInit)
	cfg.Permit(TriggerCustomerChecksBitcoin, StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateCustomerInit' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerInit' ")
		log.Println("")
	})

	// Configure States: StateBitcoinIsCheckedAndOK
	cfg = CustomerStateMachine.Configure(StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerCustomerChecksLightning, StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateBitcoinIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateLigtningIsCheckedAndOK
	cfg = CustomerStateMachine.Configure(StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateLigtningIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateCustomerWaitingForCommands
	cfg = CustomerStateMachine.Configure(StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerCommandAskForPrice, StateCustomerPriceHasBeenReceived)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerWaitingForCommands' ")
		/*if customer.askTaxiForPrice(true) == nil {
			_ = customer.askTaxiForPrice(false)
		}*/
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerWaitingForCommands' ")
		log.Println("")
	})

	// Configure States: StateCustomerPriceHasBeenReceived
	cfg = CustomerStateMachine.Configure(StateCustomerPriceHasBeenReceived)
	cfg.Permit(TriggerCustomerCommandAcceptPrice, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerPriceHasBeenReceived' ")
		/*if customer.acceptPrice(true) == nil {
			_ = customer.acceptPrice(false)
		}*/
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerPriceHasBeenReceived' ")
		log.Println("")
	})

	// Configure States: StateCustomerWaitingForPaymentRequest
	cfg = CustomerStateMachine.Configure(StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerReceivesPaymentRequest, StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerCommandHaltPayments, StateCustomerHaltedPayments)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerWaitingForPaymentRequest' ")
		if customer.PaymentStreamStarted == false {
			go receiveTaxiInvoices(customerClient, useEnvironment)
		}
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerWaitingForPaymentRequest' ")
		log.Println("")
	})

	// Configure States: StateCustomerPaymentRequestReceived
	cfg = CustomerStateMachine.Configure(StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerCustomerCommandHaltPayments, StateCustomerHaltedPayments)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerPaymentRequestReceived' ")

	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerPaymentRequestReceived' ")
		log.Println("")
	})

	// Configure States: StateCustomerHaltedPayments
	cfg = CustomerStateMachine.Configure(StateCustomerHaltedPayments)
	cfg.Permit(TriggerCustomerWillContinueToPay, StateCustomerPaymentRequestReceived)
	cfg.Permit(TriggerCustomerCommandLeaveTaxi, StateCustomerLeftTaxi)
	cfg.Permit(TriggerCustomerContiniueToWaitForPaymentRequest, StateCustomerWaitingForPaymentRequest)
	cfg.Permit(TriggerTaxiKicksOutCustomer, StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerHaltedPayments' ")
		//_ = customer.CustomerStateMachine.Fire(TriggerCustomerSetsHardwareInDriveMode.Key, nil)
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerHaltedPayments' ")
		log.Println("")
	})

	// Configure States: StateTaxiKickedOutCustomer
	cfg = CustomerStateMachine.Configure(StateTaxiKickedOutCustomer)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxiKickedOutCustomer' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxiKickedOutCustomer' ")
		log.Println("")
	})

	// Configure States: StateCustomerLeftTaxi
	cfg = CustomerStateMachine.Configure(StateCustomerLeftTaxi)
	cfg.Permit(TriggerCustomerWaitsForCommands, StateCustomerWaitingForCommands)
	cfg.Permit(TriggerCustomerEndsInErrorMode, StateCustomerIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateCustomerLeftTaxi' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerLeftTaxi' ")
		log.Println("")
	})

	// Configure States: StateCustomerIsInErrorMode
	cfg = CustomerStateMachine.Configure(StateCustomerIsInErrorMode)

	cfg.OnEnter(func() { log.Println("*** Entering 'StateCustomerIsInErrorMode' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerIsInErrorMode' ")
		log.Println("")
	})

	customer.CustomerStateMachine = CustomerStateMachine

	return customer
}

// ******************************************************************************
// log Errors for Triggers and States
func logTriggerStateError(spaceCount int, currentState ssm.State, trigger ssm.Trigger, err error) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, "Current state:", currentState, " doesn't accept trigger'", trigger, "'. Error Message: ", err)
}

// ******************************************************************************

// ******************************************************************************
// Log Errors for Triggers and States
func logMessagesWithOutError(spaceCount int, message string) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, message)
}

// ******************************************************************************
// Log Errors for Triggers and States
func logMessagesWithError(spaceCount int, message string, err error) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, message, err)
}

// ******************************************************************************
// Ask Taxi for Price
func CallBackAskTaxiForPrice(emptyParameter *server.EmptyParameter) (*server.Price_UI, error) {
	log.Println("Incoming: 'AskTaxiForPrice'")

	returnMessage := &server.Price_UI{
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

				returnMessage = &server.Price_UI{
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

func CallBackAcceptPrice(emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error){

	log.Println("Incoming: 'AcceptPrice'")

	returnMessage := &server.AckNackResponse{
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

				returnMessage = &server.AckNackResponse{
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
func CallBackHaltPayments(haltPaymentRequestMessage *server.HaltPaymentRequest) (*server.AckNackResponse, error){

	log.Println("Incoming: 'HaltPayments' with parameter: ", haltPaymentRequestMessage.Haltpayment)

	returnMessage := &server.AckNackResponse{
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

					returnMessage = &server.AckNackResponse{
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
					returnMessage = &server.AckNackResponse{
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
					returnMessage = &server.AckNackResponse{
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

func CallBackLeaveTaxi(emptyParameter *server.EmptyParameter) (*server.AckNackResponse, error){

	log.Println("Incoming: 'LeaveTaxi'")

	returnMessage := &server.AckNackResponse{
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

				returnMessage = &server.AckNackResponse{
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
// Simulate a Customer recieves paymentRequest Stream
func receiveTaxiInvoices(client taxi_grpc_api.TaxiClient, enviroment *taxi_grpc_api.Enviroment) {
	log.Printf("Starting Taxi PaymentRequest stream %v", enviroment)
	var invoices []string

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := client.PaymentRequestStream(ctx, enviroment)
	if err != nil {
		log.Fatalf("Problem to connect to Taxi Invoice Stream: ", client, err)
	}
	// Used for controlling state machine
	customer.PaymentStreamStarted = true
	for {
		if customer.CustomerStateMachine.IsInState(StateCustomerWaitingForPaymentRequest) {

			invoice, err := stream.Recv()
			if err == io.EOF {
				log.Println("HMMM, skumt borde inte slutat här när vi tar emot InvoiceStream, borde avsluta Customer")
				customer.PaymentStreamStarted = false
				break
			}
			if err != nil {
				log.Fatalf("Problem when streaming from Taxi invoice Stream:", client, err)
			}
			customer.receivedTaxiInvoiceButNotPaid = true

			//Customer Pays Invoice
			invoices = append(invoices, invoice.LightningPaymentRequest)
			log.Println("Invoices: ", invoices)

			stateChangeResponse := customer.CustomerStateMachine.Fire(TriggerCustomerReceivesPaymentRequest.Key, nil)
			if stateChangeResponse != nil {
				log.Println("Error, Should be able to change state with Trigger: 'TriggerCustomerReceivesPaymentRequest'")
			} else {

				if customer.CustomerStateMachine.IsInState(StateCustomerPaymentRequestReceived) {
					err = lightningConnection.PayReceivedInvoicesFromTaxi(invoices)
					if err != nil {
						log.Println("Problem when paying Invoice from Taxi: ", err)
						customer.PaymentStreamStarted = false
						break
					} else {
						log.Println("Invoice from Taxi-Stream is paid: ", invoices)

						invoices = append(invoices[:0], invoices[1:]...)

						customer.receivedTaxiInvoiceButNotPaid = true

					}

					stateChangeResponse = customer.CustomerStateMachine.Fire(TriggerCustomerPaysPaymentRequest.Key, nil)
					if stateChangeResponse != nil {
						log.Println("Error, Should be able to change state with Trigger: 'TriggerCustomerPaysPaymentRequest'")
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
			log.Fatalf("error creating new btc client: %v", err)
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

// ******************************************************************************
// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		remoteTaxiServerConnection.Close()

		/*
		log.Println("Gracefull stop for: 'customerRpcUIServer'")
		customerRpcUIServer.GracefulStop()
*/
		log.Println("Gracefull stop for: 'customerRpcUIStreamServer'")
		customerRpcUIStreamServer.GracefulStop()
/*
		log.Println("Close net.Listing: %v", localCustomerUIRPCServerEngineLocalPort)
		lis.Close()
*/
		log.Println("Close net.Listing: %v", localCustomerUIRPCStreamServerEngineLocalPort)
		lisStream.Close()

	}
}

func main() {

	var err error

	defer cleanup()

	// *********************
	// Set up connection to Toll Gate Hardware Server
	remoteTaxiServerConnection, err = grpc.Dial(taxi_address_to_dial, grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect to Taxi Server on address: ", taxi_address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		log.Println("gRPC connection OK to Taxi Server, address: ", taxi_address_to_dial)
		// Creates a new Clients
		customerClient = taxi_grpc_api.NewTaxiClient(remoteTaxiServerConnection)

	}

	//Init Lightning
	go lightningConnection.LigtningMainService()

	//Start Web Backend
	go gopherjs.Webmain(CallBackAskTaxiForPrice, CallBackAcceptPrice, CallBackHaltPayments, CallBackLeaveTaxi)
/*
	// *********************
	// Start Customer RPC UI Server for Incomming web connectionss
	log.Println("Customer UI RPC Server started")
	log.Println("Start listening on: %v", localCustomerUIRPCServerEngineLocalPort)
	lis, err = net.Listen("tcp", localCustomerUIRPCServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
*/
	//*****************


	// Start Customer RPC UI Stream Server for Incomming web connectionss
	log.Println("Customer UI RPC Stream Server started")
	log.Println("Start listening on: %v", localCustomerUIRPCStreamServerEngineLocalPort)
	lisStream, err = net.Listen("tcp", localCustomerUIRPCStreamServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}


/*
	// Creates a new RegisterClient gRPC server for UI
	go func() {
		log.Println("Starting Customer RPC UI Server")
		customerRpcUIServer = grpc.NewServer()
		//customerRpcUIServer = grpc.NewServer(grpc.Creds(creds))
		customer_ui_api.RegisterCustomer_UIServer(customerRpcUIServer, &customerUIServiceServer{})
		log.Println("registerTaxiServer for Taxi Gate started")
		log.Println(customerRpcUIServer.GetServiceInfo())
		customerRpcUIServer.Serve(lis)


	}()
*/

	// Creates a new RegisterClient gRPC stream server for UI
	go func() {
		log.Println("Starting Customer RPC UI Stream Server")
		customerRpcUIStreamServer = grpc.NewServer()
		customer_ui_stream_api.RegisterCustomerUIPriceStreamServer(customerRpcUIStreamServer, &customerUIPriceStreamServiceServer{})
		log.Println("registerTaxiServer for Taxi Gate started")
		customerRpcUIStreamServer.Serve(lisStream)
	}()
	// *********************



	// Set up the Private Taxi Road State Machine
	initiateCustomer()


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
		fmt.Println("sleeping...for another 5 minutes")
		time.Sleep(300 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}

}
