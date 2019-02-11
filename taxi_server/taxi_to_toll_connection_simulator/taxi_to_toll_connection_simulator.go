package main

import (
	"strings"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"log"
	"fmt"
	"os"
	taxoTotollGate_api "github.com/jlambert68/lightningCab/grpc_api/toll_road_grpc_api"
	"bufio"
	"github.com/jlambert68/lightningCab/common_config"
)

var (
	remoteServerConnection *grpc.ClientConn
	testClient             taxoTotollGate_api.TollRoadServerClient

	address_to_dial string                         = common_config.GrpcTollServer_port //"127.0.0.1:50651"
	useEnv                                         = taxoTotollGate_api.TestOrProdEnviroment_Test
	useEnvironment  *taxoTotollGate_api.Enviroment = &taxoTotollGate_api.Enviroment{TestOrProduction: useEnv}
)

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
// Simulate a Taxi that request a PaymentRequest

func requestPaymentRequest() {

	resp, err := testClient.GetPaymentRequest(context.Background(), useEnvironment)
	if err != nil {
		logMessagesWithError(4, "Could not send 'GetPaymentRequest' to address: "+address_to_dial+". Error Message:", err)

	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'GetPaymentRequest' on address "+address_to_dial+" successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.ReturnMessage)
			logMessagesWithOutError(4, "Payment Request: "+resp.GetPaymentRequest())
		} else {
			logMessagesWithOutError(4, "'GetPaymentRequest' on address "+address_to_dial+" NOT successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.ReturnMessage)
			logMessagesWithOutError(4, "Payment Request: "+resp.GetPaymentRequest())

		}
	}

}

// ******************************************************************************
// Used for only process cleanup once
var cleanupProcessed bool = false

func cleanup() {

	if cleanupProcessed == false {

		cleanupProcessed = true

		// Cleanup before close down application
		log.Println("Clean up and shut down servers")

		remoteServerConnection.Close()
	}
}

func main() {

	var err error

	defer cleanup()

	// *********************
	// Set up connection to Toll Gate Hardware Server
	remoteServerConnection, err = grpc.Dial(address_to_dial, grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect to Toll Server on address: ", address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		log.Println("gRPC connection OK to Toll Server, address: ", address_to_dial)
		// Creates a new Clients
		testClient = taxoTotollGate_api.NewTollRoadServerClient(remoteServerConnection)

	}

	fmt.Println("Enter 'GetPaymentrequest' to trigger get payment request")
	fmt.Println("")

	for {

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		switch text {
		case "GetPaymentrequest":
			requestPaymentRequest()

			fmt.Println("Please try again!")
			fmt.Println("")

		default:
			fmt.Println("Unknown command, please try again!")
			fmt.Println("")
		}
	}

}
