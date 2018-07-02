package main

import (
	"strings"

	"github.com/markdaws/simple-state-machine"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	//log "log"
	tollGateHW_api "jlambert/lightningCab/toll_road_hardware_server/toll_road_hardware_grpc_api"
	tollGate_api "jlambert/lightningCab/toll_road_server/toll_road_grpc_api"

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
	"jlambert/lightningCab/toll_road_server/lightningServer"
	"jlambert/lightningCab/common_config"
)

type Toll struct {
	Title                string
	TollRoadStateMachine *ssm.StateMachine
}

var toll *Toll

var (
	remoteServerConnection *grpc.ClientConn
	testClient             tollGateHW_api.TollHardwareClient

	address_to_dial string                     = common_config.GrpcTollHardwareServer_address + common_config.GrpcTollHardwareServer_port //"127.0.0.1:50650"
	useEnv                                     = tollGateHW_api.TestOrProdEnviroment_Test
	useEnvironment  *tollGateHW_api.Enviroment = &tollGateHW_api.Enviroment{TestOrProduction: useEnv}
)

// Global connection constants
const (
	localServerEngineLocalPort = common_config.GrpcTollServer_port //":50651"
)

var (
	registerTollRoadServer *grpc.Server
	lis                    net.Listener
)

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type tollGateServiceServer struct{}

func testTollRoadCycle() {

	toll = NewTollRoad("Test Toll Road cycle")
	//toll.RestartToaxiSystem()
	toll.TollChecksLightning()
	toll.TollChecksHardware()
	toll.SetHardwareInFirstTimeReadyMode()
	toll.TollIsReadyAndEntersWaitState()
	//toll.TaxiConnects()
	toll.TaxiRequestsPaymentRequest(false)
	//toll.TollSendsPaymentRequestToTaxi()
	toll.TaxiPaysPaymentRequest(false)
	toll.TollValidatesThatPaymentIsOK()
	toll.TaxiLeavesToll()
	toll.TollIsReadyAndEntersWaitState()
	toll.SendTollIntoErrorState()

}

func initiateTollRoad() {
	var err error
	err = nil
	toll = NewTollRoad("Private road")
	//toll.RestartToaxiSystem()

	err = validateBitcoind()
	if err != nil {
		log.Println("Couldn't check Bitcoind, exiting system!")
		os.Exit(0)
	}
	toll.TollChecksLightning()
	toll.TollChecksHardware()
	toll.SetHardwareInFirstTimeReadyMode()
	toll.TollIsReadyAndEntersWaitState()
}

var (
	// Create States
	StateTollRoadInit                    = ssm.State{Name: "StateTollRoadInit"}
	StateBitcoinIsCheckedAndOK           = ssm.State{Name: "StateBitcoinIsCheckedAndOK"}
	StateLigtningIsCheckedAndOK          = ssm.State{Name: "StateLigtningIsCheckedAndOK"}
	StateHardwareIsCheckedAndOK          = ssm.State{Name: "StateHardwareIsCheckedAndOK"}
	StateHardwareIsInFirsteTimeReadyMode = ssm.State{Name: "StateHardwareIsInFirsteTimeReadyMode"}
	StateTollIsWaitingForTaxi            = ssm.State{Name: "StateTollIsWaitingForTaxi"}
	//StateTaxiHasConnected = ssm.State{Name: "StateTaxiHasConnected"}
	StatePaymentRequestIsCreatedAndSent = ssm.State{Name: "StatePaymentRequestIsCreatedAndSent"}
	//StateWaitingForPayment = ssm.State{Name: "StateWaitingForPayment"}
	StatePaymentRequestIsPaid         = ssm.State{Name: "StatePaymentRequestIsPaid"}
	StateThanksAndOpenTollDone        = ssm.State{Name: "StateThanksAndOpenTollDone"}
	StateClosedTollAndResetedHardware = ssm.State{Name: "StateClosedTollAndResetedHardware"}
	StateTollIsInErrorMode            = ssm.State{Name: "StateTollIsInErrorMode"}

	// Create Triggers
	TriggerRestartTollSystem               = ssm.Trigger{Key: "TriggerRestartTollSystem"}
	TriggerTollChecksBitcoin               = ssm.Trigger{Key: "TriggerTollChecksBitcoin"}
	TriggerTollChecksLightning             = ssm.Trigger{Key: "TriggerTollChecksHardware"}
	TriggerTollChecksHardware              = ssm.Trigger{Key: "TriggerTollChecksHardware"}
	TriggerSetHardwareInFirstTimeReadyMode = ssm.Trigger{Key: "SetHardwareInReadyMode"}
	TriggerTollIsReadyAndEntersWaitState   = ssm.Trigger{Key: "TriggerTollIsReadyAndEntersWaitState"}
	//TriggerTaxiConnects = ssm.Trigger{Key: "TriggerTaxiConnects"}
	TriggerTaxiRequestsPaymentRequest = ssm.Trigger{Key: "TriggerTaxiRequestsPaymentRequest"}
	//TriggerTaxiReRequestsPaymentRequest = ssm.Trigger{Key: "TriggerTaxiReRequestsPaymentRequest"}
	//TriggerTollSendsPaymentRequestToTaxi = ssm.Trigger{Key: "TriggerTollSendsPaymentRequestToTaxi"}
	TriggerTaxiPaysPaymentRequest       = ssm.Trigger{Key: "TriggerTaxiPaysPaymentRequest"}
	TriggerTollValidatesThatPaymentIsOK = ssm.Trigger{Key: "TriggerTollValidatesThatPaymentIsOK"}
	TriggerTaxiLeavesToll               = ssm.Trigger{Key: "TriggerTaxiLeavesToll"}
	TriggerTollEndsInErrorMode          = ssm.Trigger{Key: "TriggerTollEndsInErrorMode"}
)

func NewTollRoad(title string) *Toll {

	toll := &Toll{Title: title}
	// Create State machine
	tollRoadStateMachine := ssm.NewStateMachine(StateTollRoadInit)

	// Configure States: StateTollRoadInit
	cfg := tollRoadStateMachine.Configure(StateTollRoadInit)
	cfg.Permit(TriggerTollChecksBitcoin, StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateTollRoadInit' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTollRoadInit' ")
		log.Println("")
	})

	// Configure States: StateBitcoinIsCheckedAndOK
	cfg = tollRoadStateMachine.Configure(StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTollChecksLightning, StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** *** Entering 'StateBitcoinIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateLigtningIsCheckedAndOK
	cfg = tollRoadStateMachine.Configure(StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTollChecksHardware, StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateLigtningIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateHardwareIsCheckedAndOK
	cfg = tollRoadStateMachine.Configure(StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerSetHardwareInFirstTimeReadyMode, StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateHardwareIsCheckedAndOK' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateHardwareIsCheckedAndOK' ")
		log.Println("")
	})

	// Configure States: StateHardwareIsInFirsteTimeReadyMode
	cfg = tollRoadStateMachine.Configure(StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTollIsReadyAndEntersWaitState, StateTollIsWaitingForTaxi)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateHardwareIsInFirsteTimeReadyMode' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateHardwareIsInFirsteTimeReadyMode' ")
		log.Println("")
	})

	// Configure States: StateTollIsWaitingForTaxi
	cfg = tollRoadStateMachine.Configure(StateTollIsWaitingForTaxi)
	cfg.Permit(TriggerTaxiRequestsPaymentRequest, StatePaymentRequestIsCreatedAndSent)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StateTollIsWaitingForTaxi' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTollIsWaitingForTaxi' ")
		log.Println("")
	})

	/*	// Configure States: StateTaxiHasConnected
		cfg = tollRoadStateMachine.Configure(StateTaxiHasConnected)
		cfg.Permit(TriggerTaxiRequestsPaymentRequest, StatePaymentRequestIsCreatedAndSent)
		cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
		cfg.OnEnter(func() { log.Println("*** Entering 'StateTaxiHasConnected' ") })
		cfg.OnExit(func() {
			log.Println("*** Exiting 'StateTaxiHasConnected' ")
			log.Println("")
		})*/

	// Configure States: StatePaymentRequestIsCreatedAndSent
	cfg = tollRoadStateMachine.Configure(StatePaymentRequestIsCreatedAndSent)
	cfg.Permit(TriggerTaxiPaysPaymentRequest, StatePaymentRequestIsPaid)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() { log.Println("*** Entering 'StatePaymentRequestIsCreatedAndSent' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StatePaymentRequestIsCreatedAndSent' ")
		log.Println("")
	})

	/*	// Configure States: StateWaitingForPayment
		cfg = tollRoadStateMachine.Configure(StateWaitingForPayment)
		cfg.Permit(TriggerTaxiPaysPaymentRequest, StatePaymentRequestIsPaid)
		cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
		cfg.OnEnter(func() { log.Println("*** Entering 'StateWaitingForPayment' ") })
		cfg.OnExit(func() {
			log.Println("*** Exiting 'StateWaitingForPayment' ")
			log.Println("")
		})*/

	// Configure States: StatePaymentRequestIsPaid
	cfg = tollRoadStateMachine.Configure(StatePaymentRequestIsPaid)
	cfg.Permit(TriggerTollValidatesThatPaymentIsOK, StateThanksAndOpenTollDone)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StatePaymentRequestIsPaid' ")
		_ = toll.TollRoadStateMachine.Fire(TriggerTollValidatesThatPaymentIsOK.Key, nil)
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StatePaymentRequestIsPaid' ")
		log.Println("")
	})

	// Configure States: StateThanksAndOpenTollDone
	cfg = tollRoadStateMachine.Configure(StateThanksAndOpenTollDone)
	cfg.Permit(TriggerTaxiLeavesToll, StateClosedTollAndResetedHardware)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateThanksAndOpenTollDone' ")
		_ = toll.TollRoadStateMachine.Fire(TriggerTaxiLeavesToll.Key, nil)
	})

	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateThanksAndOpenTollDone' ")
		log.Println("")
	})

	// Configure States: StateClosedTollAndResetedHardware
	cfg = tollRoadStateMachine.Configure(StateClosedTollAndResetedHardware)
	cfg.Permit(TriggerTollIsReadyAndEntersWaitState, StateTollIsWaitingForTaxi)
	cfg.Permit(TriggerTollEndsInErrorMode, StateTollIsInErrorMode)
	cfg.OnEnter(func() {
		log.Println("*** Entering 'StateClosedTollAndResetedHardware' ")
		toll.SetHardwareInReadyMode()
		_ = toll.TollRoadStateMachine.Fire(TriggerTollIsReadyAndEntersWaitState.Key, nil)
	})
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateClosedTollAndResetedHardware' ")
		log.Println("")
	})

	// Configure States: StateTollIsInErrorMode
	cfg = tollRoadStateMachine.Configure(StateTollIsInErrorMode)

	cfg.OnEnter(func() { log.Println("*** Entering 'StateTollIsInErrorMode' ") })
	cfg.OnExit(func() {
		log.Println("*** Exiting 'StateTollIsInErrorMode' ")
		log.Println("")
	})

	toll.TollRoadStateMachine = tollRoadStateMachine

	return toll
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
func (toll *Toll) TollCheckBitcoin(check bool) (err error) {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollChecksBitcoin

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if toll.TollRoadStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (toll *Toll) TollChecksLightning() {

	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollChecksLightning

	err := toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)
	}

}

// *****************************************************************************

// ******************************************************************************
// Check all hardware connected to the toll
func (toll *Toll) TollChecksHardware() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollChecksHardware

	err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	switch  err.(type) {
	case nil:

		//Wait for all 3 goroutines
		wg.Add(3)

		go func() {
			// Check Servo
			defer wg.Done()

			time.Sleep(5 * time.Second)
			resp, err := testClient.CheckTollGateServo(context.Background(), useEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckTollGateServo' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckTollGateServo' on address "+address_to_dial+" says Servo is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Check E-Ink Display

			defer wg.Done()

			resp, err := testClient.CheckTollEInkDisplay(context.Background(), useEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckTollEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckTollEInkDisplay' on address "+address_to_dial+" says E-Ink Display is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckTollEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Check Distance Sensor

			defer wg.Done()

			resp, err := testClient.CheckTollDistanceSensor(context.Background(), useEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckTollDistanceSensor' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollDistanceSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckTollDistanceSensor' on address "+address_to_dial+" says E-Ink Display is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckTollDistanceSensor' on address "+address_to_dial+" says E-Ink Display is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}

			}
		}()

		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}
}

// ******************************************************************************

// ******************************************************************************
// Set the hardware in first time ready mode
func (toll *Toll) SetHardwareInFirstTimeReadyMode() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerSetHardwareInFirstTimeReadyMode

	err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)

	switch  err.(type) {
	case nil:

		//Wait for all 3 goroutines
		wg.Add(3)

		go func() {
			// CLose Servo-Gate
			defer wg.Done()

			time.Sleep(5 * time.Second)

			tollGateMessage := &tollGateHW_api.TollGateServoMessage{useEnv, tollGateHW_api.TollGateCommand_CLOSE}
			resp, err := testClient.OpenCloseTollGateServo(context.Background(), tollGateMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Closed")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Send QR Code of ip to E-Ink display

			defer wg.Done()

			eInkDisplayMessage := &tollGateHW_api.EInkDisplayMessage{useEnv, tollGateHW_api.EInkMessageType_MESSSAGE_QR, "127.0.0.1", ""}
			resp, err := testClient.UseEInkDisplay(context.Background(), eInkDisplayMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display message received")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()


		go func() {
			// Check Distance Sensor that no object is infront of the sensor

			defer wg.Done()

			distanceSensorMessage := &tollGateHW_api.DistanceSensorMessage{useEnv, tollGateHW_api.DistanceSensorCommand_OBJECT_NOT_FOUND}
			resp, err := testClient.UseDistanceSensor(context.Background(), distanceSensorMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseDistanceSensor' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollDistanceSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says OK, there are no objects infront of sensor")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says distace sensor is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}

			}
		}()

		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************
// Toll enters Wait State
func (toll *Toll) TollIsReadyAndEntersWaitState() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollIsReadyAndEntersWaitState

	err := toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************
/*
// Taxi connects to the toll
func (toll *Toll) TaxiConnects() {
	var currentTrigger  ssm.Trigger

	currentTrigger = TriggerTaxiConnects

	err := toll.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TaxiStateMachine.State(), currentTrigger, err)

	}

}
*/
// ******************************************************************************

// ******************************************************************************
// Taxi requests a payment request

func (toll *Toll) TaxiRequestsPaymentRequest(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiRequestsPaymentRequest

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if toll.TollRoadStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************

// ******************************************************************************
/*
// Taxi resends a request for a paymentrequest

func (toll *Toll) TaxiReRequestsPaymentRequest() {
	var currentTrigger  ssm.Trigger

	currentTrigger = TriggerTaxiReRequestsPaymentRequest

	err := toll.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TaxiStateMachine.State(), currentTrigger, err)

	}

}
*/

// ******************************************************************************

// ******************************************************************************
/*
// Send paymentrequest to taxi

func (toll *Toll) TollSendsPaymentRequestToTaxi() {
	var currentTrigger  ssm.Trigger

	currentTrigger = TriggerTollSendsPaymentRequestToTaxi

	err := toll.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TaxiStateMachine.State(), currentTrigger, err)

	}

}
*/
// ******************************************************************************

// ******************************************************************************
// Taxi pays paymentrequest

func taxiPaysToll(check bool) (err error) {
	err = toll.TaxiPaysPaymentRequest(check)
	return err
}

func (toll *Toll) TaxiPaysPaymentRequest(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiPaysPaymentRequest

	switch check {

	case true:
		// Do a check if state machine is in correct state for triggering event
		if toll.TollRoadStateMachine.CanFire(currentTrigger.Key) == true {
			err = nil

		} else {

			err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		}

	case false:
		// Execute Trigger
		err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
		if err != nil {
			logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

		}
	}

	return err

}

// ******************************************************************************

// ******************************************************************************
// Taxi checks that payment is OK and set hardware in correct mode
func (toll *Toll) TollValidatesThatPaymentIsOK() {

	var err error
	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollValidatesThatPaymentIsOK

	err = toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)

	switch  err.(type) {
	case nil:

		//Wait for all 3 goroutines
		wg.Add(3)

		go func() {
			// CLose Servo-Gate
			defer wg.Done()

			tollGateMessage := &tollGateHW_api.TollGateServoMessage{useEnv, tollGateHW_api.TollGateCommand_OPEN}
			resp, err := testClient.OpenCloseTollGateServo(context.Background(), tollGateMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Open")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Send Thanks to E-Ink display

			defer wg.Done()

			eInkDisplayMessage := &tollGateHW_api.EInkDisplayMessage{useEnv, tollGateHW_api.EInkMessageType_MESSAGE_TEXT, "Thanks!", ""}
			resp, err := testClient.UseEInkDisplay(context.Background(), eInkDisplayMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display message received")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}
			}
		}()

		go func() {
			// Check Distance Sensor that no object is infront of the sensor

			defer wg.Done()

			distanceSensorMessage := &tollGateHW_api.DistanceSensorMessage{useEnv, tollGateHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_LEAVES}
			resp, err := testClient.UseDistanceSensor(context.Background(), distanceSensorMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'UseDistanceSensor' to address: "+address_to_dial+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckTollDistanceSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says OK, the objects has left")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					// Signal that taxi has left the toll gate
					toll.TaxiLeavesToll()

				} else {
					logMessagesWithOutError(4, "'UseDistanceSensor' on address "+address_to_dial+" says E-ditance sensor is NOT OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)

					//Set system in Error State due to malfunctioning hardware
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

				}

			}
		}()

		wg.Wait()

	default:
		// Error triggering new state
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}


}

// ******************************************************************************
// Set the hardware in ready mode
func (toll *Toll) SetHardwareInReadyMode() {

	var wg sync.WaitGroup

	//Wait for all 31 goroutines
	wg.Add(1)

	go func() {
		// CLose Servo-Gate
		defer wg.Done()

		time.Sleep(5 * time.Second)

		tollGateMessage := &tollGateHW_api.TollGateServoMessage{useEnv, tollGateHW_api.TollGateCommand_CLOSE}
		resp, err := testClient.OpenCloseTollGateServo(context.Background(), tollGateMessage)

		if err != nil {
			logMessagesWithError(4, "Could not send 'OpenCloseTollGateServo' to address: "+address_to_dial+". Error Message:", err)
			//Set system in Error State due no connection to hardware server for 'CheckTollGateServo'
			logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
			err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
		} else {

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Gate is Closed")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			} else {
				logMessagesWithOutError(4, "'OpenCloseTollGateServo' on address "+address_to_dial+" says Servo is NOT OK")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)

				//Set system in Error State due to malfunctioning hardware
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

			}
		}
	}()

	wg.Wait()

	//Wait for all 1 goroutines
	wg.Add(1)

	go func() {
		// Send QR Code of ip to E-Ink display

		defer wg.Done()

		eInkDisplayMessage := &tollGateHW_api.EInkDisplayMessage{useEnv, tollGateHW_api.EInkMessageType_MESSSAGE_QR, "127.0.0.1", ""}
		resp, err := testClient.UseEInkDisplay(context.Background(), eInkDisplayMessage)

		if err != nil {
			logMessagesWithError(4, "Could not send 'UseEInkDisplay' to address: "+address_to_dial+". Error Message:", err)
			//Set system in Error State due no connection to hardware server for 'CheckTollEInkDisplay'
			logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
			err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)
		} else {

			if resp.GetAcknack() == true {
				logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display message received")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			} else {
				logMessagesWithOutError(4, "'UseEInkDisplay' on address "+address_to_dial+" says E-Ink Display is NOT OK")
				logMessagesWithOutError(4, "Response Message: "+resp.Comments)

				//Set system in Error State due to malfunctioning hardware
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = toll.TollRoadStateMachine.Fire(TriggerTollEndsInErrorMode.Key, nil)

			}
		}
	}()

	wg.Wait()

}

// ******************************************************************************

// ******************************************************************************
// Taxi leaves toll

func (toll *Toll) TaxiLeavesToll() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiLeavesToll

	err := toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

// ******************************************************************************
// Restarts the system when Triggered

func (toll *Toll) RestartTollSystem() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerRestartTollSystem

	err := toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************
// Send System to Error State

func (toll *Toll) SendTollIntoErrorState() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTollEndsInErrorMode

	err := toll.TollRoadStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, toll.TollRoadStateMachine.State(), currentTrigger, err)

	}

}

func validateBitcoind() (err error) {

	// Check if State machine accepts State change
	err = toll.TollCheckBitcoin(true)

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

	err = toll.TollCheckBitcoin(false)
	if err != nil {
		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)
	}

	return err
}

// Taxi request PaymentRequest
func (s *tollGateServiceServer) GetPaymentRequest(ctx context.Context, environment *tollGate_api.Enviroment) (*tollGate_api.PaymentRequestMessage, error) {

	log.Println("Incoming: 'GetPaymentRequest'")
	fmt.Println("sleeping...for 3 seconds")
	time.Sleep(3 * time.Second)

	var acknack bool
	var returnMessage string
	var paymentReqest string

	// Check if State machine accepts State change
	err := toll.TaxiRequestsPaymentRequest(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case tollGate_api.TestOrProdEnviroment_Test:
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

		case tollGate_api.TestOrProdEnviroment_Production:
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

		err = toll.TaxiRequestsPaymentRequest(false)
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

	return &tollGate_api.PaymentRequestMessage{acknack, paymentReqest, returnMessage}, nil

}

// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		log.Println("Gracefull stop for: registerTollRoadServer")
		registerTollRoadServer.GracefulStop()

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
	// Set up connection to Toll Gate Hardware Server
	remoteServerConnection, err = grpc.Dial(address_to_dial, grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect to Toll Gate Hardware Server on address: ", address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		log.Println("gRPC connection OK to Toll Gate Hardware Server, address: ", address_to_dial)
		// Creates a new Clients
		testClient = tollGateHW_api.NewTollHardwareClient(remoteServerConnection)

	}

	// *********************
	// Start Toll Gate Server for Incomming Taxi connectionss
	log.Println("Toll Gate Hardware Server started")
	log.Println("Start listening on: %v", localServerEngineLocalPort)
	lis, err = net.Listen("tcp", localServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Creates a new RegisterClient gRPC server
	go func() {
		log.Println("Starting Toll Gate Hardware Server")
		registerTollRoadServer = grpc.NewServer()
		tollGate_api.RegisterTollRoadServerServer(registerTollRoadServer, &tollGateServiceServer{})
		log.Println("registerTollRoadServer for Toll Gate started")
		registerTollRoadServer.Serve(lis)
	}()
	// *********************

	// Test State Machine in Test Mode
	useEnv = tollGateHW_api.TestOrProdEnviroment_Test
	//testTollRoadCycle()

	//Initiate Lightning
	//lightningConnection.InitLndServerConnection()
	//lightningConnection.RetrieveGetInfo()

	go lightningServer.LigtningMainService(taxiPaysToll)


	// Set up the Private Toll Road State Machine
	initiateTollRoad()


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
