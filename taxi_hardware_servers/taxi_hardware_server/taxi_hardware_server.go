package main

import (
	"fmt"
	"github.com/jlambert68/lightningCab/common_config"
	"github.com/stianeikeland/go-rpio"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	taxiHW_api "github.com/jlambert68/lightningCab/grpc_api/taxi_hardware_grpc_api"
)

// Global connection constants
const (
	localServerEngineLocalPort = common_config.GrpcTaxiHardwareServer_port
)

// Constants used for Db-objects to save
/*const (
	clientAddressInfo_DB = "clientAddressInfo_DB"
	clientTestCommands_DB = "clientTestCommands_DB"
	clientTestEnviroments_DB = "clientTestEnviroments_DB"
	clientTestInstructionBlocks_DB = "clientTestInstructionBlocks_DB"
)*/

// Variables holding all clients for gRPC-connection
//var (
//	hardwareStateMachineClient taxiHW_api.TaxiHardwareClient
//)

var (
	registerTaxiHardwareServer *grpc.Server
	lis                        net.Listener
)

var (
	// Use physical gpio pin 19 to the right of 13 that work for LED-test
	pin = rpio.Pin(19)
)

/*var (
	DB_session *mgo.Session
)*/

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type taxiHardwareServer struct{}

// Check that Power Sensor is working
func (s *taxiHardwareServer) CheckPowerSensor(ctx context.Context, environment *taxiHW_api.Enviroment) (*taxiHW_api.AckNackResponse, error) {
	log.Printf("Incoming: 'CheckPowerSensor'")

	var returnMessage string

	// Check if to Simulate or not
	switch environment.TestOrProduction {

	case taxiHW_api.TestOrProdEnviroment_Test:
		// Simulate test of Power sensor
		log.Printf("Simulate Test of Power Sensor:")
		returnMessage = "A simulated Test of the Power Sensor was done"

	case taxiHW_api.TestOrProdEnviroment_Production:
		// Use Test Taxi hardware
		log.Printf("Test of Power Sensor hardware:")
		// CALL TO HARDWARE

		returnMessage = "A test of the Power Sensor hardware was done"
	}

	return &taxiHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

// Check that Power Cutter is working
func (s *taxiHardwareServer) CheckPowerCutter(ctx context.Context, environment *taxiHW_api.Enviroment) (*taxiHW_api.AckNackResponse, error) {
	log.Printf("Incoming: 'CheckPowerCutter'")

	var returnMessage string

	log.Printf("TestOrProdEnviroment: ", environment.TestOrProduction)

	// Check if to Simulate or not
	switch environment.TestOrProduction {

	case taxiHW_api.TestOrProdEnviroment_Test:
		// Simulate test of Pwer Cutter
		log.Printf("Simulate Test of Power Cutter:")
		returnMessage = "A simulated Test of the Power Cutter was done"

	case taxiHW_api.TestOrProdEnviroment_Production:
		// Use Test Taxi hardware
		log.Printf("Test of Power Cutter hardware, pin.High")
		// CALL TO HARDWARE
		pin.High()

		log.Printf("Sleep for 3 seconds")
		time.Sleep(3 * time.Second)

		log.Printf("Test of Power Cutter hardware, pin.Low()")
		pin.Low()

		returnMessage = "A test of the Power Cutter hardware was done"
	}

	return &taxiHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

func (s *taxiHardwareServer) CutPower(ctx context.Context, powerCutterMessage *taxiHW_api.PowerCutterMessage) (*taxiHW_api.AckNackResponse, error) {
	log.Printf("Incoming: 'CheckPowerCutter'")

	var acknack bool
	var returnMessage string

	// Check if to Simulate or not
	switch powerCutterMessage.TollGateServoEnviroment {

	case taxiHW_api.TestOrProdEnviroment_Test:

		log.Printf("TestOrProdEnviroment: ", powerCutterMessage.TollGateServoEnviroment)

		switch powerCutterMessage.PowerCutterCommand {

		case taxiHW_api.PowerCutterCommand_CutPower:
			log.Printf("Simulate that Power Cutter 'Cuts' power:")
			acknack = true
			returnMessage = "A simulated Test of that Power Cutter 'Cuts' power"

		case taxiHW_api.PowerCutterCommand_HavePower:
			log.Printf("Simulate that Power Cutter 'Have' power:")
			acknack = true
			returnMessage = "A simulated Test of that Power Cutter 'Have' power"
		}

	case taxiHW_api.TestOrProdEnviroment_Production:
		// Use Test Taxi hardware
		log.Printf("TestOrProdEnviroment: ", powerCutterMessage.TollGateServoEnviroment)

		switch powerCutterMessage.PowerCutterCommand {

		case taxiHW_api.PowerCutterCommand_CutPower:
			//Cut Power on pin
			pin.Low()

			log.Printf("Execute Hardware Power Cutter 'Cuts' power:")
			acknack = true
			returnMessage = "Execute Hardware that Power Cutter 'Cuts' power"

		case taxiHW_api.PowerCutterCommand_HavePower:
			//Have Power on pin
			pin.High()

			log.Printf("Execute Hardware Power Cutter 'Have' power:")
			acknack = true
			returnMessage = "Execute Hardware that Power Cutter 'Have' power"
		}
	}

	return &taxiHW_api.AckNackResponse{Acknack: acknack, Comments: returnMessage}, nil

}



// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		log.Println("Gracefull stop for: registerTaxiHardwareServer")
		registerTaxiHardwareServer.GracefulStop()

		log.Println("Close net.Listing: %v", localServerEngineLocalPort)
		lis.Close()

		//log.Println("Close DB_session: %v", DB_session)
		//DB_session.Close()
	}
}

func main() {

	var err error

	defer cleanup()

	//*******************************
	// initate PowerCutter
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()

	//*******************************

	// Set DB Connection
	/*DB_session, err = db_func.ConnectToFenixDB()

	if err != nil {
		fmt.Println(err)
	}*/

	log.Println("Taxi Hardware Server started")
	log.Println("Start listening on: %v", localServerEngineLocalPort)
	lis, err = net.Listen("tcp", localServerEngineLocalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	/*
		// Creates a new Environment gRPC server
		go func() {
			environmentServer = grpc.NewServer()
			inception_sp.RegisterTestCommandsServer(environmentServer, &testEnvironmentServer{})
			log.Println("environmentServer for Fenix Inception started")
			environmentServer.Serve(lis)

		}()
	*/
	// Creates a new RegisterTaxiHardwareServer gRPC server
	go func() {
		log.Println("Starting Taxi Hardware Server")
		registerTaxiHardwareServer = grpc.NewServer()
		taxiHW_api.RegisterTaxiHardwareServer(registerTaxiHardwareServer, &taxiHardwareServer{})
		log.Println("registerTaxiHardwareServer for Taxi Hardware started")
		registerTaxiHardwareServer.Serve(lis)
	}()

	// Register reflection service on gRPC server.
	//reflection.Register(s)
	//if err := s.Serve(lis); err != nil {
	//	log.Fatalf("failed to serve: %v", err)
	//}

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

	//Wait until user exit
	/*
   for {
	   time.Sleep(10)
   }
   */
}
