package main

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/jlambert68/lightningCab/common_config"
	taxi_api "github.com/jlambert68/lightningCab/grpc_api/taxi_grpc_api"
	taxiHW_api "github.com/jlambert68/lightningCab/grpc_api/taxi_hardware_grpc_api"
	taxiHW_stream_api "github.com/jlambert68/lightningCab/grpc_api/taxi_hardware_grpc_stream_api"
	"github.com/markdaws/simple-state-machine"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"jlambert/lightningCab/taxi_server/lightningConnection"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Taxi struct {
	Title            string
	TaxiStateMachine *ssm.StateMachine
	logger *logrus.Logger
	invoiceStatistics invoiceStatistics_struct
}

type invoiceStatistics_struct struct {
	invoicedCreated int32
	invoicesPaid int32
}

var taxi *Taxi

var (
	remoteTaxiHWServerConnection *grpc.ClientConn
	taxiHWClient                 taxiHW_api.TaxiHardwareClient

	remoteTaxiHWStreamServerConnection *grpc.ClientConn
	taxiHWStreamClient                 taxiHW_stream_api.TaxiStreamHardwareClient

	addressToDialToTaxiHWServer       string = common_config.GrpcTaxiHardwareServer_address + common_config.GrpcTaxiHardwareServer_port
	addressToDialToTaxiHWStreamServer string = common_config.GrpcTaxiHardwareStreamServer_address + common_config.GrpcTaxiHardwareStreamServer_port

	//HW
	useEnv                                        = taxiHW_api.TestOrProdEnviroment_Production
	useHardwareEnvironment *taxiHW_api.Enviroment = &taxiHW_api.Enviroment{TestOrProduction: useEnv}

	//HW-Stream
	useEnv_stream                                                  = taxiHW_stream_api.TestOrProdEnviroment_Production
	messasurePowerMessage *taxiHW_stream_api.MessasurePowerMessage = &taxiHW_stream_api.MessasurePowerMessage{ TollGateServoEnviroment: useEnv_stream, Intervall: common_config.MilliSecondsBetweenPaymentRequest}

)

// Global connection constants
const (
	localTaxiServerEngineLocalPort = common_config.GrpcTaxiServer_port
)

var (
	registerTaxiServer *grpc.Server
	lis                net.Listener
)

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type taxiServiceServer struct{}

/*
func testTaxiCycle() {

	taxi = NewTaxi("Test Taxi Road cycle")
	//taxi.RestartToaxiSystem()
	taxi.TaxiChecksLightning()
	taxi.TaxiChecksHardware()
	taxi.SetHardwareInFirstTimeReadyMode()
	taxi.TaxiIsReadyAndEntersWaitState()
	//taxi.TaxiConnects()
	taxi.TaxiRequestsPaymentRequest(false)
	//taxi.TollSendsPaymentRequestToTaxi()
	taxi.TaxiPaysPaymentRequest(false)
	taxi.TollValidatesThatPaymentIsOK()
	taxi.TaxiLeavesToll()
	taxi.TaxiIsReadyAndEntersWaitState()
	taxi.SendTaxiIntoErrorState()

}*/

func initTaxi() {
	taxi = NewTaxi("MyTaxi")
}

func (taxi *Taxi) startTaxi() {
	var err error
	err = nil
	//taxi = NewTaxi("MyTaxi")
	//taxi.RestartToaxiSystem()

	err = validateBitcoind()
	if err != nil {
		//log.Println("Couldn't check Bitcoind, exiting system!")
		taxi.logger.Error("Couldn't check Bitcoind, exiting system!")

		os.Exit(0)
	}
	taxi.TaxiChecksLightning()
	taxi.TaxiChecksHardware()
	taxi.SetHardwareInFirstTimeReadyMode()
	taxi.TaxiIsReadyAndEntersWaitState()
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
	TriggerTaxiChecksLightning                = ssm.Trigger{Key: "TriggerTaxiChecksHardware"}
	TriggerTaxiChecksHardware                 = ssm.Trigger{Key: "TriggerTaxiChecksHardware"}
	TriggerSetHardwareInFirstTimeReadyMode    = ssm.Trigger{Key: "TriggerSetHardwareInFirstTimeReadyMode"}
	TriggerTaxiIsReadyAndEntersWaitState      = ssm.Trigger{Key: "TriggerTaxiIsReadyAndEntersWaitState"}
	TriggerCustomerConnects                   = ssm.Trigger{Key: "TriggerCustomerConnects"}
	TriggerCustomerAcceptsPrice               = ssm.Trigger{Key: "TriggerCustomerAcceptsPrice"}
	TriggerTaxiSetsHardwareInDriveMode        = ssm.Trigger{Key: "TriggerTaxiSetsHardwareInDriveMode"}
	TriggerTaxiStopsStreamsAndWaitsforPayment = ssm.Trigger{Key: "TriggerTaxiStopsStreamsAndWaitsforPayment"}
	TriggerCustomerPaysPaymentRequest         = ssm.Trigger{Key: "TriggerCustomerPaysPaymentRequest"}
	TriggerCustomerStoppsPayingTimer          = ssm.Trigger{Key: "TriggerCustomerStoppsPayingTimer"}
	TriggerCustomerLeavesTaxi                 = ssm.Trigger{Key: "TriggerCustomerLeavesTaxi"}
	TriggerTaxiResetsHardwareForNewCustomer   = ssm.Trigger{Key: "TriggerTaxiResetsHardwareForNewCustomer"}
	TriggerTaxiEndsInErrorMode                = ssm.Trigger{Key: "TriggerTaxiEndsInErrorMode"}
)

func NewTaxi(title string) *Taxi {

	taxi := &Taxi{
		Title: title,
		invoiceStatistics: invoiceStatistics_struct{invoicedCreated:0 , invoicesPaid: 0},
	}
	// Create State machine
	taxiStateMachine := ssm.NewStateMachine(StateTaxiInit)

	// Configure States: StateTaxiInit
	cfg := taxiStateMachine.Configure(StateTaxiInit)
	cfg.Permit(TriggerTaxiChecksBitcoin, StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** *** Entering 'StateTaxiInit' ")
		taxi.logger.Info("xxxx*** *** Entering 'StateTaxiInit' xxxx")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxiInit' ")
		taxi.logger.Info("*** Exiting 'StateTaxiInit' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateBitcoinIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateBitcoinIsCheckedAndOK)
	cfg.Permit(TriggerTaxiChecksLightning, StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** *** Entering 'StateBitcoinIsCheckedAndOK' ")
		taxi.logger.Info("*** *** Entering 'StateBitcoinIsCheckedAndOK' ")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		taxi.logger.Info("*** Exiting 'StateBitcoinIsCheckedAndOK' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateLigtningIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateLigtningIsCheckedAndOK)
	cfg.Permit(TriggerTaxiChecksHardware, StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateLigtningIsCheckedAndOK' ")
		taxi.logger.Info("*** Entering 'StateLigtningIsCheckedAndOK' ")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		taxi.logger.Info("*** Exiting 'StateLigtningIsCheckedAndOK' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateHardwareIsCheckedAndOK
	cfg = taxiStateMachine.Configure(StateHardwareIsCheckedAndOK)
	cfg.Permit(TriggerSetHardwareInFirstTimeReadyMode, StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateHardwareIsCheckedAndOK' ")
		taxi.logger.Info("*** Entering 'StateHardwareIsCheckedAndOK' ")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateHardwareIsCheckedAndOK' ")
		taxi.logger.Info("*** Exiting 'StateHardwareIsCheckedAndOK' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateHardwareIsInFirsteTimeReadyMode
	cfg = taxiStateMachine.Configure(StateHardwareIsInFirsteTimeReadyMode)
	cfg.Permit(TriggerTaxiIsReadyAndEntersWaitState, StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateHardwareIsInFirsteTimeReadyMode' ")
		taxi.logger.Info("*** Entering 'StateHardwareIsInFirsteTimeReadyMode' ")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateHardwareIsInFirsteTimeReadyMode' ")
		taxi.logger.Info("*** Exiting 'StateHardwareIsInFirsteTimeReadyMode' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateTaxisWaitingForCustomer
	cfg = taxiStateMachine.Configure(StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerCustomerConnects, StateCustomerHasConnected)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateTaxisWaitingForCustomer' ")
		taxi.logger.Info("*** Entering 'StateTaxisWaitingForCustomer' ")
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxisWaitingForCustomer' ")
		taxi.logger.Info("*** Exiting 'StateTaxisWaitingForCustomer' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateCustomerHasConnected
	cfg = taxiStateMachine.Configure(StateCustomerHasConnected)
	cfg.Permit(TriggerCustomerAcceptsPrice, StateCustomerHasAcceptedPrice)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateCustomerHasConnected' ")
		taxi.logger.Info("*** Entering 'StateCustomerHasConnected' ")
	})
		cfg.OnExit(func() {
			//log.Println("*** Exiting 'StateCustomerHasConnected' ")
			taxi.logger.Info("*** Exiting 'StateCustomerHasConnected' ")
			//log.Println("")
			taxi.logger.Info("")
		})

	// Configure States: StateCustomerHasAcceptedPrice
	cfg = taxiStateMachine.Configure(StateCustomerHasAcceptedPrice)
	cfg.Permit(TriggerTaxiSetsHardwareInDriveMode, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerHasAcceptedPrice' ")
		taxi.logger.Info("*** Entering 'StateCustomerHasAcceptedPrice' ")
		_ = taxi.TaxiStateMachine.Fire(TriggerTaxiSetsHardwareInDriveMode.Key, nil)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerHasAcceptedPrice' ")
		taxi.logger.Info("*** Exiting 'StateCustomerHasAcceptedPrice' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateTaxiIsReadyToDrive
	cfg = taxiStateMachine.Configure(StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerTaxiStopsStreamsAndWaitsforPayment, StateTaxiIsWaitingForPayment)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateTaxiIsReadyToDrive' ")
		taxi.logger.Info("*** Entering 'StateTaxiIsReadyToDrive' ")
		// Have Power to Engine
		//taxi.CutPowerToEngine(false)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxiIsReadyToDrive' ")
		taxi.logger.Info("*** Exiting 'StateTaxiIsReadyToDrive' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateTaxiIsWaitingForPayment
	cfg = taxiStateMachine.Configure(StateTaxiIsWaitingForPayment)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerCustomerStoppsPayingTimer, StateCustomerStoppedPaying)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() { //log.Println("*** Entering 'StateTaxiIsWaitingForPayment' ")
		taxi.logger.Info("*** Entering 'StateTaxiIsWaitingForPayment' ")

	})
		cfg.OnExit(func() {
			//log.Println("*** Exiting 'StateTaxiIsWaitingForPayment' ")
			taxi.logger.Info("*** Exiting 'StateTaxiIsWaitingForPayment' ")
			//log.Println("")
			taxi.logger.Info("")
		})

	// Configure States: StateCustomerStoppedPaying
	cfg = taxiStateMachine.Configure(StateCustomerStoppedPaying)
	cfg.Permit(TriggerCustomerPaysPaymentRequest, StateTaxiIsReadyToDrive)
	cfg.Permit(TriggerCustomerLeavesTaxi, StateCustomerLeftTaxi)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerStoppedPaying' ")
		taxi.logger.Info("*** Entering 'StateCustomerStoppedPaying' ")
		_ = taxi.TaxiStateMachine.Fire(TriggerCustomerLeavesTaxi.Key, nil)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerStoppedPaying' ")
		taxi.logger.Info("*** Exiting 'StateCustomerStoppedPaying' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateCustomerLeftTaxi
	cfg = taxiStateMachine.Configure(StateCustomerLeftTaxi)
	cfg.Permit(TriggerTaxiResetsHardwareForNewCustomer, StateTaxiIsReadyForNewCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateCustomerLeftTaxi' ")
		// Cut Power to Engine
		taxi.CutPowerToEngine(true)
		taxi.logger.Info("*** Entering 'StateCustomerLeftTaxi' ")

		_ = taxi.SetHardwareInNextCustomerReadyMode(false)
	})

	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateCustomerLeftTaxi' ")
		taxi.logger.Info("*** Exiting 'StateCustomerLeftTaxi' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateTaxiIsReadyForNewCustomer
	cfg = taxiStateMachine.Configure(StateTaxiIsReadyForNewCustomer)
	cfg.Permit(TriggerTaxiIsReadyAndEntersWaitState, StateTaxisWaitingForCustomer)
	cfg.Permit(TriggerTaxiEndsInErrorMode, StateTaxiIsInErrorMode)
	cfg.OnEnter(func() {
		//log.Println("*** Entering 'StateTaxiIsReadyForNewCustomer' ")
		taxi.logger.Info("*** Entering 'StateTaxiIsReadyForNewCustomer' ")
		_ = taxi.TaxiStateMachine.Fire(TriggerTaxiIsReadyAndEntersWaitState.Key, nil)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxiIsReadyForNewCustomer' ")
		taxi.logger.Info("*** Exiting 'StateTaxiIsReadyForNewCustomer' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	// Configure States: StateTaxiIsInErrorMode
	cfg = taxiStateMachine.Configure(StateTaxiIsInErrorMode)

	cfg.OnEnter(func() { //log.Println("*** Entering 'StateTaxiIsInErrorMode' ")
		taxi.logger.Info("*** Entering 'StateTaxiIsInErrorMode' ")
		// Cut Power to Engine
		taxi.CutPowerToEngine(true)
	})
	cfg.OnExit(func() {
		//log.Println("*** Exiting 'StateTaxiIsInErrorMode' ")
		taxi.logger.Info("*** Exiting 'StateTaxiIsInErrorMode' ")
		//log.Println("")
		taxi.logger.Info("")
	})

	taxi.TaxiStateMachine = taxiStateMachine

	return taxi
}

// ******************************************************************************
// log Errors for Triggers and States
func logTriggerStateError(spaceCount int, currentState ssm.State, trigger ssm.Trigger, err error) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, "Current state:", currentState, " doesn't accept trigger'", trigger, "'. Error Message: ", err)
	messageToLog := "Current state: " + currentState.Name + " doesn't accept trigger'" + trigger.String() + "'."
	taxi.logger.Info(messageToLog)
}

// ******************************************************************************
// log Errors for Triggers and States
func logMessagesWithOutError(spaceCount int, message string) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, message)
	taxi.logger.Info(message)
}

// ******************************************************************************
// log Errors for Triggers and States
func logMessagesWithError(spaceCount int, message string, err error) {

	//spaces := strings.Repeat("  ", spaceCount)
	//log.Println(spaces, message, err)
	taxi.logger.WithFields(logrus.Fields{
		"error": err,
	}).Info(message)
}

// ******************************************************************************
// Check Bitcoind and Lightning Nodes
func (taxi *Taxi) TaxiCheckBitcoin(check bool) (err error) {

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
func (taxi *Taxi) TaxiChecksLightning() {

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
			resp, err := taxiHWClient.CheckPowerSensor(context.Background(), useHardwareEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckPowerSensor' to address: "+addressToDialToTaxiHWServer+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckPowerSensor'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckPowerSensor' on address "+addressToDialToTaxiHWServer+" says Power Sensor is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckPowerSensor' on address "+addressToDialToTaxiHWServer+" says Power Sensor is NOT OK")
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

			resp, err := taxiHWClient.CheckPowerCutter(context.Background(), useHardwareEnvironment)
			if err != nil {
				logMessagesWithError(4, "Could not send 'CheckPowerCutter' to address: "+addressToDialToTaxiHWServer+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'CheckPowerCutter'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'CheckPowerCutter' on address "+addressToDialToTaxiHWServer+" says Power Cutter is OK")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'CheckPowerCutter' on address "+addressToDialToTaxiHWServer+" says Power Cutter is NOT OK")
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

			PowerCutterMessage := &taxiHW_api.PowerCutterMessage{TollGateServoEnviroment:useEnv, PowerCutterCommand:taxiHW_api.PowerCutterCommand_CutPower}
			resp, err := taxiHWClient.CutPower(context.Background(), PowerCutterMessage)

			if err != nil {
				logMessagesWithError(4, "Could not send 'PowerCutterMessage' to address: "+addressToDialToTaxiHWServer+". Error Message:", err)
				//Set system in Error State due no connection to hardware server for 'PowerCutterMessage'
				logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
				err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
			} else {

				if resp.GetAcknack() == true {
					logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says Gate is Closed")
					logMessagesWithOutError(4, "Response Message: "+resp.Comments)
				} else {
					logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says Servo is NOT OK")
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
// Set the hardware in ready mode for Next Cusotmer
func (taxi *Taxi) SetHardwareInNextCustomerReadyMode(check bool) (err error) {

	var wg sync.WaitGroup
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiResetsHardwareForNewCustomer

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

		switch  err.(type) {
		case nil:

			//Wait for all 1 goroutine
			wg.Add(1)

			go func() {
				// Cut Power
				defer wg.Done()

				time.Sleep(5 * time.Second)

				PowerCutterMessage := &taxiHW_api.PowerCutterMessage{TollGateServoEnviroment: useEnv,PowerCutterCommand:taxiHW_api.PowerCutterCommand_CutPower}
				resp, err := taxiHWClient.CutPower(context.Background(), PowerCutterMessage)

				if err != nil {
					logMessagesWithError(4, "Could not send 'PowerCutterMessage' to address: "+addressToDialToTaxiHWServer+". Error Message:", err)
					//Set system in Error State due no connection to hardware server for 'PowerCutterMessage'
					logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
					err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
				} else {

					if resp.GetAcknack() == true {
						logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says Gate is Closed")
						logMessagesWithOutError(4, "Response Message: "+resp.Comments)
					} else {
						logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says Servo is NOT OK")
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
	return err
}

// ******************************************************************************

// ******************************************************************************
// Taxi enters Wait State
func (taxi *Taxi) TaxiIsReadyAndEntersWaitState() {
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
// Customer Stops paying

func (taxi *Taxi) PaymentsStopsComing(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiStopsStreamsAndWaitsforPayment

	// Allways Stop Taxi Engine if we come here
	//Stop Engine
	//taxi.CutPowerToEngine(true)


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
// Customer continues to pay invoice-stream after short stop

func (taxi *Taxi) continueStreamingPaymentRequests(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerPaysPaymentRequest

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

func (taxi *Taxi) abortPaymentRequestGeneration(check bool) (err error) {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerCustomerStoppsPayingTimer

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
// Cut or UnCut Power to Engine

func (taxi *Taxi) CutPowerToEngine(cutPower bool) {
	var PowerCutterMessage= &taxiHW_api.PowerCutterMessage{}


	if cutPower == true {
		PowerCutterMessage = &taxiHW_api.PowerCutterMessage{TollGateServoEnviroment: useEnv, PowerCutterCommand: taxiHW_api.PowerCutterCommand_CutPower}
		logMessagesWithOutError(4, "Send: 'PowerCutterCommand_CutPower'")
	} else {
		PowerCutterMessage = &taxiHW_api.PowerCutterMessage{TollGateServoEnviroment: useEnv, PowerCutterCommand: taxiHW_api.PowerCutterCommand_HavePower}
		logMessagesWithOutError(4, "Send: 'PowerCutterCommand_HavePower'")
	}
	resp, err := taxiHWClient.CutPower(context.Background(), PowerCutterMessage)

	if err != nil {
		logMessagesWithError(4, "Could not send 'PowerCutterMessage' to address: "+addressToDialToTaxiHWServer+". Error Message:", err)
		//Set system in Error State due no connection to hardware server for 'PowerCutterMessage'
		logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
		err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)
	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says OK")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
		} else {
			logMessagesWithOutError(4, "'PowerCutterMessage' on address "+addressToDialToTaxiHWServer+" says PowerCutter is NOT OK")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)

			//Set system in Error State due to malfunctioning hardware
			logMessagesWithOutError(4, "Putting State machine into Error state and Stop")
			err = taxi.TaxiStateMachine.Fire(TriggerTaxiEndsInErrorMode.Key, nil)

		}
	}
}


// ******************************************************************************
// Restarts the system when Triggered

func (taxi *Taxi) RestartToaxiSystem() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerRestartTaxiSystem

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************
// Send System to Error State

func (taxi *Taxi) SendTaxiIntoErrorState() {
	var currentTrigger ssm.Trigger

	currentTrigger = TriggerTaxiEndsInErrorMode

	err := taxi.TaxiStateMachine.Fire(currentTrigger.Key, nil)
	if err != nil {
		logTriggerStateError(4, taxi.TaxiStateMachine.State(), currentTrigger, err)

	}

}

// ******************************************************************************

func validateBitcoind() (err error) {

	// Check if State machine accepts State change
	err = taxi.TaxiCheckBitcoin(true)

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
			taxi.logger.WithFields(logrus.Fields{
				"err":    err,
			}).Fatal("Error creating new btc client")

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

	err = taxi.TaxiCheckBitcoin(false)
	if err != nil {
		logMessagesWithError(4, "State machine is not in correct state to be able to check Bitcoind: ", err)
	}

	return err
}

// ******************************************************************************
// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		//log.Println("Clean up and shut down servers")
		taxi.logger.Info("Clean up and shut down servers")

		//log.Println("Gracefull stop for: registerTaxiServer")
		taxi.logger.Info("Gracefull stop for: registerTaxiServer")
		registerTaxiServer.GracefulStop()

		//log.Println("Close net.Listing: %v", localTaxiServerEngineLocalPort)
		taxi.logger.WithFields(logrus.Fields{
			"localTaxiServerEngineLocalPort":    localTaxiServerEngineLocalPort,
		}).Info("Close net.Listing")
		lis.Close()

		//log.Println("Close DB_session: %v", DB_session)
		//DB_session.Close()
		remoteTaxiHWServerConnection.Close()
		remoteTaxiHWStreamServerConnection.Close()
	}
}

//var log = logging.MustGetLogger("")

func main() {

	var err error

	defer cleanup()

	// Init Customer Object
	initTaxi()

	// Init logger
	taxi.InitLogger("")

	// Should only be done from init functions
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(taxi.logger.Out, taxi.logger.Out, taxi.logger.Out))


	//initLog()

	// *********************
	// Set up connection to Taxi Hardware Server
	remoteTaxiHWServerConnection, err = grpc.Dial(addressToDialToTaxiHWServer, grpc.WithInsecure())
	if err != nil {
		//log.Println("did not connect to Taxi Hardware Server on address: ", addressToDialToTaxiHWServer, "error message", err)
		taxi.logger.WithFields(logrus.Fields{
			"addressToDialToTaxiHWServer":    addressToDialToTaxiHWServer,
			"error message": err,
		}).Error("did not connect to Taxi Hardware Server on address")
		os.Exit(0)
	} else {
		//log.Println("gRPC connection OK to Taxi  Hardware Server, address: ", addressToDialToTaxiHWServer)
		taxi.logger.WithFields(logrus.Fields{
			"addressToDialToTaxiHWServer":    addressToDialToTaxiHWServer,
		}).Info("gRPC connection OK to Taxi  Hardware Server")
		// Creates a new Clients
		taxiHWClient = taxiHW_api.NewTaxiHardwareClient(remoteTaxiHWServerConnection)

	}

	// *********************
	// Set up connection to Taxi Hardware Stream Server
	remoteTaxiHWStreamServerConnection, err = grpc.Dial(addressToDialToTaxiHWStreamServer, grpc.WithInsecure())
	if err != nil {
		//log.Println("did not connect to Taxi Hardware Stream Server on address: ", addressToDialToTaxiHWStreamServer, "error message", err)
		taxi.logger.WithFields(logrus.Fields{
			"addressToDialToTaxiHWStreamServer":    addressToDialToTaxiHWStreamServer,
			"error message": err,
		}).Error("did not connect to Taxi Hardware Stream Server on address")
		os.Exit(0)
	} else {
		//log.Println("gRPC connection OK to Taxi Hardware Stream Server, address: ", addressToDialToTaxiHWStreamServer)
		taxi.logger.WithFields(logrus.Fields{
			"addressToDialToTaxiHWStreamServer":    addressToDialToTaxiHWStreamServer,
		}).Info("gRPC connection OK to Taxi Hardware Stream Server")
		// Creates a new Clients
		taxiHWStreamClient = taxiHW_stream_api.NewTaxiStreamHardwareClient(remoteTaxiHWStreamServerConnection)

	}

	// Starting to receive stream from Taxi Engine about Powerconsumption
	go receiveEnginePowerdata(taxiHWStreamClient, messasurePowerMessage)

	// *********************
	// Start Taxi Server for Incomming Customer connectionss
	//log.Println("Taxi Customer Server started")
	taxi.logger.Info("Taxi Customer Server started")
	//log.Println("Start listening on: %v", localTaxiServerEngineLocalPort)
	taxi.logger.WithFields(logrus.Fields{
		"localTaxiServerEngineLocalPort":    localTaxiServerEngineLocalPort,
	}).Info("Start listening on")
	lis, err = net.Listen("tcp", localTaxiServerEngineLocalPort)
	if err != nil {
		//log.Fatalf("failed to listen: %v", err)
		taxi.logger.WithFields(logrus.Fields{
			"error":    err,
		}).Fatal("failed to listen:")
	}

	// Creates a new RegisterClient gRPC server
	go func() {
		//log.Println("Starting Taxi Customer Server")
		taxi.logger.Info("Starting Taxi Customer Server")
		registerTaxiServer = grpc.NewServer()
		taxi_api.RegisterTaxiServer(registerTaxiServer, &taxiServiceServer{})
		//log.Println("registerTaxiServer for Taxi Gate started")
		taxi.logger.Info("registerTaxiServer for Taxi Gate started")
		registerTaxiServer.Serve(lis)
	}()
	// *********************

	// Test State Machine in Test/Production Mode
	useEnv = taxiHW_api.TestOrProdEnviroment_Production
	//testTaxiCycle()

	//Initiate Lightning
	//lightningConnection.InitLndServerConnection()
	//lightningConnection.RetrieveGetInfo()

	go lightningConnection.LigtningMainService(taxi.logger, customerPaysPaymentRequest)

	// Set up the Private Taxi Road State Machine
	taxi.startTaxi()

	// Set system in wait mode for externa input, Taxi and payment ...
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	for {
		//fmt.Println("sleeping...for another 5 minutes")
		taxi.logger.Info("sleeping...for another 5 minutes")

		time.Sleep(300 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}

}
