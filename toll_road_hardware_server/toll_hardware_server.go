package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	tollGateHW_api "jlambert/lightningCab/toll_road_hardware_server/toll_road_hardware_grpc_api"
)

// Global connection constants
const (
	localServerEngineLocalPort = ":50650"
)

// Constants used for Db-objects to save
/*const (
	clientAddressInfo_DB = "clientAddressInfo_DB"
	clientTestCommands_DB = "clientTestCommands_DB"
	clientTestEnviroments_DB = "clientTestEnviroments_DB"
	clientTestInstructionBlocks_DB = "clientTestInstructionBlocks_DB"
)*/

// Variables holding all clients for gRPC-connection
var (
	hardwareStateMachineClient tollGateHW_api.TollHardwareClient
)

var (
	registerTollHardwareServer *grpc.Server
	lis                        net.Listener
)

/*var (
	DB_session *mgo.Session
)*/

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type tollGateHardwareServiceServer struct{}

// Check that Servo for Toll Gate is working
func (s *tollGateHardwareServiceServer) CheckTollGateServo(ctx context.Context, environment *tollGateHW_api.Enviroment) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'CheckTollGateServo'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch environment.TestOrProduction {

	case tollGateHW_api.TestOrProdEnviroment_Test:
		// Simulate test of Toll Gate Servo
		log.Printf("Simulate Test of Toll Gate Servo:")
		returnMessage = "A simulated Test of the Toll Gate Servo was done"

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use Test Toll Gate Servo hardware
		log.Printf("Test of Toll Gate Servo hardware:")
		// CALL TO HARDWARE
		returnMessage = "A test of the Toll Gate Servo hardware was done"
	}

	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

//Check that Distance sensor is working
func (s *tollGateHardwareServiceServer) CheckTollDistanceSensor(ctx context.Context, environment *tollGateHW_api.Enviroment) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'CheckTollDistanceSensor'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch environment.TestOrProduction {

	case tollGateHW_api.TestOrProdEnviroment_Test:
		// Simulate test of Toll Gate Servo
		log.Printf("Simulate Test of Distance Sensor:")
		returnMessage = "A simulated Test of the Distance Sensor was done"

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use Test Toll Gate Servo hardware
		log.Printf("Test of Distance Sensor hardware:")
		// CALL TO HARDWARE
		returnMessage = "A test of the TDistance Sensor hardware was done"
	}

	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

//Check that E-ink display is working
func (s *tollGateHardwareServiceServer) CheckTollEInkDisplay(ctx context.Context, environment *tollGateHW_api.Enviroment) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'CheckTollEInkDisplay'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch environment.TestOrProduction {

	case tollGateHW_api.TestOrProdEnviroment_Test:
		// Simulate test of Toll Gate Servo
		log.Printf("Simulate Test of E-Ink Display:")
		returnMessage = "A simulated Test of the E-Ink Display was done"

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use Test Toll Gate Servo hardware
		log.Printf("Test of E-Ink Display hardware:")
		// CALL TO HARDWARE
		returnMessage = "A test of the E-Ink Display hardware was done"
	}

	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

// ************************************
// Open or Close the the Toll Gate
// ************************************
func (s *tollGateHardwareServiceServer) OpenCloseTollGateServo(ctx context.Context, tollGateServoMsg *tollGateHW_api.TollGateServoMessage) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'OpenCloseTollGateServo'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch tollGateServoMsg.TollGateServoEnviroment {

	case tollGateHW_api.TestOrProdEnviroment_Test:

		// Simulate Open/Close Toll Gate
		switch tollGateServoMsg.TollGateCommand {

		case tollGateHW_api.TollGateCommand_OPEN:
			// Simulate Open Gate
			log.Printf("Simulate Open Toll Gate Servo:")
			returnMessage = "A simulated opening of the gate was done"

		case tollGateHW_api.TollGateCommand_CLOSE:
			//Simulate Close Gate
			log.Printf("Simulate Close Toll Gate Servo:")
			returnMessage = "A simulated closing of the gate was done"

		}

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use hardware to Open/Close Toll Gate
		switch tollGateServoMsg.TollGateCommand {

		case tollGateHW_api.TollGateCommand_OPEN:
			// Open Gate
			log.Printf("Open Toll Gate Servo:")
			// CALL TO HARDWARE
			returnMessage = "The gate was opened"
		case tollGateHW_api.TollGateCommand_CLOSE:
			//Close Gate
			log.Printf("Close Toll Gate Servo:")
			// CALL TO HARDWARE
			returnMessage = "The gate was closed"
		}
	}

	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

// ************************************
// Use the Distance Sensor
// ************************************
func (s *tollGateHardwareServiceServer) UseDistanceSensor(ctx context.Context, distanceSensorMsg *tollGateHW_api.DistanceSensorMessage) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'UseDistanceSensor'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch distanceSensorMsg.DistanceSensorEnviroment {

	case tollGateHW_api.TestOrProdEnviroment_Test:
		// Simulate Distance Sensor
		switch distanceSensorMsg.DistanceSensorCommand {
		case tollGateHW_api.DistanceSensorCommand_OBJECT_FOUND:
			log.Printf("Simulate Object infront of sensor:")
			returnMessage = "A simulate object was found infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_OBJECT_NOT_FOUND:
			log.Printf("Simulate No Object infront of sensor:")
			returnMessage = "A simulate object was not found infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_ENTERS:
			log.Printf("Simulate Object enters infront of sensor:")
			returnMessage = "A simulate object has entered infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_LEAVES:
			log.Printf("Simulate Object leaves space infront of sensor:")
			returnMessage = "A simulate object has left infront of sensor"
		}

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use Distance Sensor
		switch distanceSensorMsg.DistanceSensorCommand {

		case tollGateHW_api.DistanceSensorCommand_OBJECT_FOUND:
			log.Printf("Check sensor if Object is infront of sensor:")
			//Call Hardware
			returnMessage = "An object was found infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_OBJECT_NOT_FOUND:
			log.Printf("Check sensor if no Object is infront of sensor:")
			//Call Hardware
			returnMessage = "An object was not found infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_ENTERS:
			log.Printf("Sensor signal when object enters infront of sensor:")
			//Call Hardware
			returnMessage = "An object has entered infront of sensor"

		case tollGateHW_api.DistanceSensorCommand_SIGNAL_WHEN_OBJECT_LEAVES:
			log.Printf("Sensor signal when object leaves space infront of sensor:")
			//Call Hardware
			returnMessage = "An object has left infront of sensor"
		}
	}

	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

// ************************************
// Send Message to E-Ink display
// ************************************
func (s *tollGateHardwareServiceServer) UseEInkDisplay(ctx context.Context, eInkDisplayMsg *tollGateHW_api.EInkDisplayMessage) (*tollGateHW_api.AckNackResponse, error) {

	log.Printf("Incoming: 'UseEInkDisplay'")
	fmt.Println("sleeping...for 10 seconds")
	time.Sleep(10 * time.Second)

	var returnMessage string

	// Check if to Simulate or not
	switch eInkDisplayMsg.EInkDisplayEnviroment {

	case tollGateHW_api.TestOrProdEnviroment_Test:
		// Simulate E-Ink display
		switch eInkDisplayMsg.MessageType {

		case tollGateHW_api.EInkMessageType_MESSAGE_TEXT:
			log.Printf("Simulate Send Text to E-Ink display:")

		case tollGateHW_api.EInkMessageType_MESSSAGE_QR:
			log.Printf("Simulate Send QC-code to E-Ink display:")

		case tollGateHW_api.EInkMessageType_MESSAGE_PICTURE:
			log.Printf("Simulate Send Picture to E-Ink display:")

		}

	case tollGateHW_api.TestOrProdEnviroment_Production:
		// Use E-Ink display
		switch eInkDisplayMsg.MessageType {

		case tollGateHW_api.EInkMessageType_MESSAGE_TEXT:
			log.Printf("Send Text to E-Ink display:")
			//Call Hardware

		case tollGateHW_api.EInkMessageType_MESSSAGE_QR:
			log.Printf("Send QC-code to E-Ink display:")
			//Call Hardware

		case tollGateHW_api.EInkMessageType_MESSAGE_PICTURE:
			log.Printf("Send Picture to E-Ink display:")
			//Call Hardware

		}
	}

	returnMessage = "THe message was sent to the E-Ink display"
	return &tollGateHW_api.AckNackResponse{Acknack: true, Comments: returnMessage}, nil

}

// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		log.Println("Gracefull stop for: registerTollHardwareServer")
		registerTollHardwareServer.GracefulStop()

		log.Println("Close net.Listing: %v", localServerEngineLocalPort)
		lis.Close()

		//log.Println("Close DB_session: %v", DB_session)
		//DB_session.Close()
	}
}

func main() {

	var err error

	defer cleanup()

	// Set DB Connection
	/*DB_session, err = db_func.ConnectToFenixDB()

	if err != nil {
		fmt.Println(err)
	}*/

	log.Println("Toll Gate Hardware Server started")
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
	// Creates a new RegisterTollHardwareServer gRPC server
	go func() {
		log.Println("Starting Toll Gate Hardware Server")
		registerTollHardwareServer = grpc.NewServer()
		tollGateHW_api.RegisterTollHardwareServer(registerTollHardwareServer, &tollGateHardwareServiceServer{})
		log.Println("registerTollHardwareServer for Toll Gate Hardware started")
		registerTollHardwareServer.Serve(lis)
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
