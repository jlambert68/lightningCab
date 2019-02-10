package main

import (
	"fmt"
	"github.com/jlambert68/lightningCab/common_config"
	"google.golang.org/grpc"
	//"jlambert/lightningCab/common_config"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	//"golang.org/x/net/context"

	taxiHW_stream_api "jlambert/lightningCab/taxi_hardware_servers/taxi_hardware_server_stream/taxi_hardware_grpc_stream_api"
)

// Global connection constants
const (
	localServerEngineLocalPort = common_config.GrpcTaxiHardwareStreamServer_port
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
//	hardwareStateMachineClient taxiHW_stream_api.TaxiHardwareClient
//)

var (
	registerTaxiHardwareStreamServer *grpc.Server
	lis                              net.Listener
)

/*var (
	DB_session *mgo.Session
)*/

// Server used for register clients Name, Ip and Por and Clients Test Enviroments and Clients Test Commandst
type taxiHardwareStreamServer struct{}

func (s *taxiHardwareStreamServer) MessasurePowerConsumption(messasurePowerMessage *taxiHW_stream_api.MessasurePowerMessage, stream taxiHW_stream_api.TaxiStreamHardware_MessasurePowerConsumptionServer) (err error) {
	log.Printf("Incoming: 'MessasurePowerConsumption'")

	err = nil
	powerConsumption := &taxiHW_stream_api.PowerStatusResponse{
		Acknack: true,
		Comments: "Standard return message",
		Speed: 0,
		Acceleration: 0,
		Timestamp: time.Now().UnixNano()}

	for {
		if err := stream.Send(powerConsumption); err != nil {
			return err
			log.Printf("Error when streaming back: 'MessasurePowerConsumption'")
			break
		}
		//log.Printf("Sent the following powerdata: ", powerConsumption)
		time.Sleep(100 * time.Millisecond)

		powerConsumption.Speed = powerConsumption.Speed + 1
		if powerConsumption.Speed > 100 {
			powerConsumption.Speed = 0
			//log.Println("Powerconsuption: ", powerConsumption)
		}
		powerConsumption.Acceleration = powerConsumption.Acceleration + 2
		if powerConsumption.Acceleration > 100 {
			powerConsumption.Acceleration = 0
			//log.Println("Powerconsuption: ", powerConsumption)
		}
		now := time.Now()
		powerConsumption.Timestamp = now.UnixNano()
	}

	log.Println("Leaving stream service!")
	return nil
}

// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		log.Println("Gracefull stop for: registerTaxiHardwareStreamServer")
		registerTaxiHardwareStreamServer.GracefulStop()

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

	log.Println("Taxi Hardware Stream Server started")
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
		registerTaxiHardwareStreamServer = grpc.NewServer()
		taxiHW_stream_api.RegisterTaxiStreamHardwareServer(registerTaxiHardwareStreamServer, &taxiHardwareStreamServer{})
		log.Println("registerTaxiHardwareStreamServer for Taxi Hardware started")
		registerTaxiHardwareStreamServer.Serve(lis)
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
