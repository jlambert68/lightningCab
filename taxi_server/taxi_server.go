package main

import (
	"strings"

	"github.com/markdaws/simple-state-machine"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	//log "log"
	taxiHW_api "jlambert/lightningCab/taxi_hardware_server/taxi_hardware_grpc_api"
	taxi_api "jlambert/lightningCab/taxi_server/taxi_grpc_api"

	"time"
	"sync"
	"net"
	"os"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"strconv"

	//"github.com/op/go-logging"
	"log"
	"jlambert/lightningCab/taxi_server/lightningServer"
	"jlambert/lightningCab/common_config"
)

type Taxi struct {
	Title            string
	TaxiStateMachine *ssm.StateMachine
}

var taxi *Taxi

var (
	remoteServerConnection *grpc.ClientConn
	testClient             taxiHW_api.TaxiHardwareClient

	address_to_dial string                 = common_config.GrpcTaxiHardwareServer_address + common_config.GrpcTaxiHardwareServer_port
	useEnv                                 = taxiHW_api.TestOrProdEnviroment_Test
	useEnvironment  *taxiHW_api.Enviroment = &taxiHW_api.Enviroment{TestOrProduction: useEnv}
)

// Global connection constants
const (
	localServerEngineLocalPort = common_config.GrpcTaxiServer_port
)

var (
	registerTaxiServer *grpc.Server
	lis                net.Listener
)

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type taxiServiceServer struct{}

func testTaxiCycle() {

	taxi = NewTaxi("Test Taxi Road cycle")
	//taxi.RestartTollSystem()
	taxi.TollChecksLightning()
	taxi.TaxiChecksHardware()
	taxi.SetHardwareInFirstTimeReadyMode()
	taxi.TollIsReadyAndEntersWaitState()
	//taxi.TaxiConnects()
	taxi.TaxiRequestsPaymentRequest(false)
	//taxi.TollSendsPaymentRequestToTaxi()
	taxi.TaxiPaysPaymentRequest(false)
	taxi.TollValidatesThatPaymentIsOK()
	taxi.TaxiLeavesToll()
	taxi.TollIsReadyAndEntersWaitState()
	taxi.SendTollIntoErrorState()

}

func initiateTaxi() {
	var err error
	err = nil
	taxi = NewTaxi("MyTaxi")
	//taxi.RestartTollSystem()

	err = validateBitcoind()
	if err != nil {
		log.Println("Couldn't check Bitcoind, exiting system!")
		os.Exit(0)
	}
	taxi.TollChecksLightning()
	taxi.TaxiChecksHardware()
	taxi.SetHardwareInFirstTimeReadyMode()
	taxi.TollIsReadyAndEntersWaitState()
}

var (
	// Create States
	StateTaxiInit                        = ssm.State{Name: "StateTaxiInit"}
	StateBitcoinIsCheckedAndOK           = ssm.State{Name: "StateBitcoinIsCheckedAndOK"}
	StateLigtningIsCheckedAndOK          = ssm.State{Name: "StateLigtningIsCheckedAndOK"}
	StateHardwareIsCheckedAndOK          = ssm.State{Name: "StateHardwareIsCheckedAndOK"}
	StateHardwareIsInFirsteTimeReadyMode = ssm.State{Name: "StateHardwareIsInFirsteTimeReadyMode"}
	StateTaxisWaitingForCustomer         = ssm.State{Name: "StateTaxisWaitingForCustomer"}
	StateCustomerHasConnected            = ssm.State{Name: "StateCustomerHasConnected"}
	StateCustomerHasAcceptedPrice        = ssm.State{Name: "StateCustomerHasAcceptedPrice"}
	StateTaxiIsReadyToDrive              = ssm.State{Name: "StateTaxiIsReadyToDrive"}
	StateTaxiIsWaitingForPayment         = ssm.State{Name: "StateTaxiIsWaitingForPayment"}
	StateCustomerStoppedPaying           = ssm.State{Name: "StateCustomerStoppedPaying"}
	StateCustomerLeftTaxi                = ssm.State{Name: "StateCustomerLeftTaxi"}
	StateTaxiIsReadyForNewCustomer       = ssm.State{Name: "StateTaxiIsReadyForNewCustomer"}
	StateTaxiIsInErrorMode               = ssm.State{Name: "StateTaxiIsInErrorMode"}

	// Create Triggers
	TriggerRestartTaxiSystem                   = ssm.Trigger{Key: "TriggerRestartTaxiSystem"}
	TriggerTaxiChecksBitcoin                   = ssm.Trigger{Key: "TriggerTaxiChecksBitcoin"}
	TriggerTaxiChecksLightning                 = ssm.Trigger{Key: "TriggerTaxiChecksHardware"}
	TriggerTaxiChecksHardware                  = ssm.Trigger{Key: "TriggerTaxiChecksHardware"}
	TriggerSetHardwareInFirstTimeReadyMode     = ssm.Trigger{Key: "TriggerSetHardwareInFirstTimeReadyMode"}
	TriggerTaxiIsReadyAndEntersWaitState       = ssm.Trigger{Key: "TriggerTaxiIsReadyAndEntersWaitState"}
	TriggerCustomerConnects                    = ssm.Trigger{Key: "TriggerCustomerConnects"}
	TriggerCustomerAcceptsPrice                = ssm.Trigger{Key: "TriggerCustomerAcceptsPrice"}
	TriggerTaxiSetsHardwareInDriveMode         = ssm.Trigger{Key: "TriggerTaxiSetsHardwareInDriveMode"}
	TriggerTaxiStreamsPaymentRequestToCustomer = ssm.Trigger{Key: "TriggerTaxiStreamsPaymentRequestToCustomer"}
	TriggerCustomerPaysPaymentRequest          = ssm.Trigger{Key: "TriggerCustomerPaysPaymentRequest"}
	TriggerCustomerStoppsPayingTimer           = ssm.Trigger{Key: "TriggerCustomerStoppsPayingTimer"}
	TriggerCustomerLeavesTaxi                  = ssm.Trigger{Key: "TriggerCustomerLeavesTaxi"}
	TriggerTaxiResetsHardware                  = ssm.Trigger{Key: "TriggerCustomerLeavesTaxi"}
	TriggerTaxiEndsInErrorMode                 = ssm.Trigger{Key: "TriggerTaxiEndsInErrorMode"}
)

func NewTaxi(title string) *Taxi {

	taxi := &Taxi{Title: title}
	// Create State machine
	taxiStateMachine := ssm.NewStateMachine(StateTaxiInit)

	// Configure States: StateTaxiInit
	cfg := taxiStateMachine.Configure(StateTaxiInit)
	cfg.Permit(TriggerTaxiChecksBitcoin, StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateTaxiInit' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxiInit' ")
		log.Println("")
	})

	// Configure States: StateBitcoinIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTaxiChecksLightning, StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateBitcoinIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateLigtningIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTaxiChecksHardware, StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateLigtningIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateHardwareIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerSetHardwareInFirstTimeReadyMode, StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateHardwareIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateHardwareIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateHardwareIsInFirsteTimeReadyMode
	cfg = taxiStateMachine.Configure(StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTaxiIsReadyAndEntersWaitState, StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateHardwareIsInFirsteTimeReadyMode' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateHardwareIsInFirsteTimeReadyMode' ")
		log.Println("")
	})

	// Configure States: StateTaxisWaitingForCustomer
	cfg = taxiStateMachine.Configure(StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerCustomerConnects, StateCustomerHasConnected)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxisWaitingForCustomer' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxisWaitingForCustomer' ")
		log.Println("")
	})

	// Configure States: StateCustomerHasConnected
	cfg = taxiStateMachine.Configure(StateCustomerHasConnected)
	cfg.Permit(TriggerCustomerAcceptsPrice, StateCustomerHasAcceptedPrice)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateCustomerHasConnected' ") })
		cfg.OnExit(func() {
			log.Println("*** Exiting 'StateCustomerHasConnected' ")
			log.Println("")
		})

	// Configure States: StateCustomerHasAcceptedPrice
	cfg = taxiStateMachine.Configure(StateCustomerHasAcceptedPrice)
	cfg.Permit(TriggerTaxiSetsHardwareInDriveMode, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateCustomerHasAcceptedPrice' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerHasAcceptedPrice' ")
		log.Println("")
	})

	// Configure States: StateTaxiIsReadyToDrive
	cfg = taxiStateMachine.Configure(StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerTaxiStreamsPaymentRequestToCustomer, StateTaxiIsWaitingForPayment)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxiIsReadyToDrive' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxiIsReadyToDrive' ")
		log.Println("")
	})

	// Configure States: StateTaxiIsWaitingForPayment
	cfg = taxiStateMachine.Configure(StateTaxiIsWaitingForPayment)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerCustomerStoppsPayingTimer, StateCustomerStoppedPaying)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxiIsWaitingForPayment' ") })
		cfg.OnExit(func() {
			log.Println("*** Exiting 'StateTaxiIsWaitingForPayment' ")
			log.Println("")
		})

	// Configure States: StateCustomerStoppedPaying
	cfg = taxiStateMachine.Configure(StateCustomerStoppedPaying)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerCustomerLeavesTaxi, StateCustomerLeftTaxi)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerStoppedPaying' ")
		//_ = taxi.TaxiStateMachine.Fire(TriggerTollValidatesThatPaymentIsOK.Key, nil)
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerStoppedPaying' ")
		log.Println("")
	})

	// Configure States: StateCustomerLeftTaxi
	cfg = taxiStateMachine.Configure(StateCustomerLeftTaxi)
	cfg.Permit(TriggerTaxiResetsHardware, StateTaxiIsReadyForNewCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateCustomerLeftTaxi' ")
		//_ = taxi.TaxiStateMachine.Fire(TriggerTaxiLeavesToll.Key, nil)
	})

	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateCustomerLeftTaxi' ")
		log.Println("")
	})

	// Configure States: StateTaxiIsReadyForNewCustomer
	cfg = taxiStateMachine.Configure(StateTaxiIsReadyForNewCustomer)
	cfg.Permit(TriggerTaxiIsReadyAndEntersWaitState, StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateTaxiIsReadyForNewCustomer' ")
		taxi.SetHardwareInReadyMode()
		//_ = taxi.TaxiStateMachine.Fire(TriggerTaxiIsReadyAndEntersWaitState.Key, nil)
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxiIsReadyForNewCustomer' ")
		log.Println("")
	})

	// Configure States: StateTaxiIsInErrorMode
	cfg = taxiStateMachine.Configure(StateTaxiIsInErrorMode)

	cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxiIsInErrorMode' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTaxiIsInErrorMode' ")
		log.Println("")
	})

	taxi.TaxiStateMachine = taxiStateMachine

	return taxi
}

// ******************************************************************************
// log Errors for Triggers and States
func logTriggerStateError(spaceCount int, currentState ssm.State, trigger ssm.Trigger, err error) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, "Current state:", currentState, " doesn't accept trigger'", trigger, "'. Error Message: ", err)
}

// ******************************************************************************
// log Errors for Triggers and States
func logMessagesWithOutError(spaceCount int, message string) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, message)
}

// ******************************************************************************
// log Errors for Triggers and States
func logMessagesWithError(spaceCount int, message string, err error) {

	spaces := strings.Repeat("  ", spaceCount)
	log.Println(spaces, message, err)
}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (taxi *Taxi) TollCheckBitcoin(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiChecksBitcoin

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if taxi.TaxiStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (taxi *Taxi) TollChecksLightning() {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiChecksLightning

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)
	}

}

// *****************************************************************************

// ******************************************************************************
// Check all hardware connected to the taxi
func (taxi *Taxi) TaxiChecksHardware() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiChecksHardware

	err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	switch  err.(type) {
	case nil:

		//Wait for all 3 goroutines
		wg.Add(2)

		go func() {
			// Check Power Sensor
			defer wg.Done()

			time.Sleep(5 * time.Second)
			resp, err := testClient.CheckPowerSensor(context.Background(), useEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckPowerSensor' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckPowerSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckPowerSensor' on address "+address_to_dial+" says Power Sensor is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckPowerSensor' on address "+address_to_dial+" says Power Sensor is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Check Power Cutter

			defer wg.Done()

			resp, err := testClient.CheckPowerCutter(context.Background(), useEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckPowerCutter' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckPowerCutter'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckPowerCutter' on address "+address_to_dial+" says Power Cutter is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckPowerCutter' on address "+address_to_dial+" says Power Cutter is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}
			}
		}()

		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}
}

// ******************************************************************************

// ******************************************************************************
// Set the hardware in first time ready mode
func (taxi *Taxi) SetHardwareInFirstTimeReadyMode() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerSetHardwareInFirstTimeReadyMode

	err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)

	switch  err.(type) {
	case nil:

		//Wait for all 1 goroutine
		wg.Add(1)

		go func() {
			// Cut Power
			defer wg.Done()

			time.Sleep(5 * time.Second)

			PowerCutterMessage := &taxiHW_api.PowerCutterMessage{useEnv, taxiHW_api.PowerCutterCommand_CutPower}
			resp, err := testClient.CutPower(context.Background(), PowerCutterMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Closed")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}
			}
		}()


		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************
// Taxi enters Wait State
func (taxi *Taxi) TollIsReadyAndEntersWaitState() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiIsReadyAndEntersWaitState

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************

// Customer connects to the taxi
func (taxi *Taxi) CustomerConnects(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerConnects

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if taxi.TaxiStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Customer Accepts Price
func (taxi *Taxi) CustomerAcceptsPrice(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerAcceptsPrice

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if taxi.TaxiStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}
// ******************************************************************************

// ******************************************************************************
// Taxi requests a payment request

func (taxi *Taxi) TaxiRequestsPaymentRequest(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiStreamsPaymentRequestToCustomer

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if taxi.TaxiStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************

// ******************************************************************************
/*
// Taxi resends a request for a paymentrequest

func (taxi *Taxi) TaxiReRequestsPaymentRequest() {
	var currentTrigger  ssm.Trigger

	currentTrigger = TriggerTaxiReRequestsPaymentRequest

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}
*/

// ******************************************************************************

// ******************************************************************************
/*
// Send paymentrequest to taxi

func (taxi *Taxi) TollSendsPaymentRequestToTaxi() {
	var currentTrigger  ssm.Trigger

	currentTrigger = TriggerTollSendsPaymentRequestToTaxi

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}
*/
// ******************************************************************************

// ******************************************************************************
// Taxi pays paymentrequest

func taxiPaysToll(check bool) (err error) {
	err = taxi.TaxiPaysPaymentRequest(check)
	return err
}

func (taxi *Taxi) TaxiPaysPaymentRequest(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiPaysPaymentRequest

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if taxi.TaxiStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************

// ******************************************************************************
// Taxi checks that payment is OK and set hardware in correct mode
func (taxi *Taxi) TollValidatesThatPaymentIsOK() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollValidatesThatPaymentIsOK

	err = taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)

	switch  err.(type) {
	case nil:

		//Wait for all 3 goroutines
		wg.Add(3)

		go func() {
			// CLose Servo-Gate
			defer wg.Done()

			tollGateMessage := &taxiHW_api.TollGateServoMessage{useEnv, taxiHW_api.TollGateCommand_OPEN}
			resp, err := testClient.OpenCloseTollGateServo(context.Background(), tollGateMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Open")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Send Thanks to E-Ink display

			defer wg.Done()

			eInkDisplayMessage := &taxiHW_api.EInkDisplayMessage{useEnv, taxiHW_api.EInkMessageType_MESSAGE_TEXT, "Thanks!", ""}
			resp, err := testClient.UseEInkDisplay(context.Background(), eInkDisplayMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display message received")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Check Distance Sensor that no object is infront of the sensor

			defer wg.Done()

			distanceSensorMessage := &taxiHW_api.DistanceSensorMessage{useEnv, taxiHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_LEAVES}
			resp, err := testClient.UseDistanceSensor(context.Background(), distanceSensorMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseDistanceSensor' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollDistanceSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says OK, the objects has left")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					// Signal that taxi has left the taxi gate
					taxi.TaxiLeavesToll()

				} else {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says E-ditance sensor is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

				}

			}
		}()

		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************
// Set the hardware in ready mode
func (taxi *Taxi) SetHardwareInReadyMode() {

	var wg sync.WaitGroup

	//Wait for all 31 goroutines
	wg.Add(1)

	go func() {
		// CLose Servo-Gate
		defer wg.Done()

		time.Sleep(5 * time.Second)

		tollGateMessage := &taxiHW_api.TollGateServoMessage{useEnv, taxiHW_api.TollGateCommand_CLOSE}
		resp, err := testClient.OpenCloseTollGateServo(context.Background(), tollGateMessage)

		if err != nil {
			logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
			//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
			logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
			err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
		} else {

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Closed")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			} else {
				logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)

				//Set system in Error State due to malfunctioning hardware
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

			}
		}
	}()

	wg.Wait()

	//Wait for all 1 goroutines
	wg.Add(1)

	go func() {
		// Send QR Code of ip to E-Ink display

		defer wg.Done()

		eInkDisplayMessage := &taxiHW_api.EInkDisplayMessage{useEnv, taxiHW_api.EInkMessageType_MESSSAGE_QR, "127.0.0.1", ""}
		resp, err := testClient.UseEInkDisplay(context.Background(), eInkDisplayMessage)

		if err != nil {
			logMessagesWithError(4, "Could not send 'UseEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
			//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
			logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
			err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
		} else {

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display message received")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			} else {
				logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)

				//Set system in Error State due to malfunctioning hardware
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

			}
		}
	}()

	wg.Wait()

}

// ******************************************************************************

// ******************************************************************************
// Taxi leaves taxi

func (taxi *Taxi) TaxiLeavesToll() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiLeavesToll

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************
// Restarts the system when Triggered

func (taxi *Taxi) RestartTollSystem() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerRestartTaxiSystem

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************
// Send System to Error State

func (taxi *Taxi) SendTollIntoErrorState() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiEndsInErrorMode

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

func validateBitcoind() (err error) {

	// Check if State machine accepts State change
	err = taxi.TollCheckBitcoin(true)

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
	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)

	}

	err = taxi.TollCheckBitcoin(false)
	if err != nil {
		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)
	}

	return err
}

// Taxi gererates a PaymentRequest
func GeneratePaymentRequest(ctx context.Context, environment *taxi_api.Enviroment) (*taxi_api.PaymentRequestMessage, error) {

	log.Println("Incoming: 'GetPaymentRequest'")


	var acknack bool
	var returnMessage string
	var paymentReqest string

	// Check if State machine accepts State change
	err := taxi.TaxiRequestsPaymentRequest(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case taxi_api.TestOrProdEnviroment_Test:
			// Simulate send payment request to Taxi
			log.Println("Create Payment Request to Taxi:")
			//acknack = true
			//returnMessage = "Payment Request Created"
			//paymentReqest = "kjdfhksdjfhlskdjfhlasdkjfhsldkjfhksdfjhlkdsjfhdskjfhskdjfhsk"

			invoice, err := lightningServer.CreateInvoice("Payment Request for Taxi", 100, 180)
			if err != nil || invoice.Invoice == "" {
				acknack = false
				returnMessage = "There was a problem when creating payment request"
				paymentReqest = ""
			} else {
				acknack = true
				returnMessage = "Payment Request Created"
				paymentReqest = invoice.Invoice
			}

		case taxi_api.TestOrProdEnviroment_Production:
			// Send payment request to Taxi
			log.Println("Create Payment Request to Taxi:")
			// Create Payment Request
			//acknack = true
			//returnMessage = "Payment Request Created"
			//paymentReqest = "kjdfhksdjfhlskdjfhlasdkjfhsldkjfhksdfjhlkdsjfhdskjfhskdjfhsk"

			invoice, err := lightningServer.CreateInvoice("Payment Request for Taxi", 100, 180)
			if err != nil || invoice.Invoice == "" {
				acknack = false
				returnMessage = "There was a problem when creating payment request"
				paymentReqest = ""
			} else {
				acknack = true
				returnMessage = "Payment Request Created"
				paymentReqest = invoice.Invoice
			}

		default:
			logMessagesWithOutError(4, "Unknown incomming enviroment: "+environment.TestOrProduction.String())
			acknack = false
			returnMessage = "Unknown incomming enviroment: " + environment.TestOrProduction.String()
			paymentReqest = ""
		}

		err = taxi.TaxiRequestsPaymentRequest(false)
		if err != nil {
			acknack = false
			returnMessage = "There was a problem when creating payment request"
			paymentReqest = ""
		} else {
			invoice, err := lightningServer.CreateInvoice("Payment Request for Taxi", 100, 180)
			if err != nil || invoice.Invoice == "" {
				acknack = false
				returnMessage = "There was a problem when creating payment request"
				paymentReqest = ""
			}
		}

	} else {

		logMessagesWithError(4, "State machine is not in correct state: ", err)
		acknack = false
		returnMessage = "State machine is not in correct state to be able to create payment request"
		paymentReqest = ""
	}

	return &taxi_api.PaymentRequestMessage{acknack, paymentReqest, returnMessage}, nil

}

// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		log.Println("Gracefull stop for: registerTaxiServer")
		registerTaxiServer.GracefulStop()

		log.Println("Close net.Listing: %v", localServerEngineLocalPort)
		lis.Close()

		//log.Println("Close DB_session: %v", DB_session)
		//DB_session.Close()
		remoteServerConnection.Close()
	}
}

//var log = logging.MustGetLogger("")

func main() {

	var err error

	defer cleanup()

	//initLog()

	// *********************
	// Set up connection to Taxi Gate Hardware Server
	remoteServerConnection, err = grpc.Dial(address_to_dial, grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect to Taxi Gate Hardware Server on address: ", address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		log.Println("gRPC connection OK to Taxi Gate Hardware Server, address: ", address_to_dial)
		// Creates a new Clients
		testClient = taxiHW_api.NewTollHardwareClient(remoteServerConnection)

	}

	// *********************
	// Start Taxi Gate Server for Incomming Taxi connectionss
	log.Println("Taxi Gate Hardware Server started")
	log.Println("Start listening on: %v", localServerEngineLocalPort)
	lis, err = net.Listen("tcp", localServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Creates a new RegisterClient gRPC server
	go func() {
		log.Println("Starting Taxi Gate Hardware Server")
		registerTaxiServer = grpc.NewServer()
		taxi_api.RegisterTollRoadServerServer(registerTaxiServer, &taxiServiceServer{})
		log.Println("registerTaxiServer for Taxi Gate started")
		registerTaxiServer.Serve(lis)
	}()
	// *********************

	// Test State Machine in Test Mode
	useEnv = taxiHW_api.TestOrProdEnviroment_Test
	//testTaxiCycle()

	//Initiate Lightning
	//lightningServer.InitLndServerConnection()
	//lightningServer.RetrieveGetInfo()

	go lightningServer.LigtningMainService(taxiPaysToll)

	// Set up the Private Taxi Road State Machine
	initiateTaxi()

	// Set system in wait mode for externa input, Taxi and payment ...
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
