package main

import (
	"strings"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"log"
	"fmt"
	"os"
	"bufio"
	"jlambert/lightningCab/common_config"
	xxx "github.com/jlambert68/lightningCab/grpc_api/customer_grpc_ui_api"
)

var (
	remoteServerConnection *grpc.ClientConn
	testClient             xxx.Customer_UIClient

	address_to_dial string                           = common_config.GrpcCustomerUI_RPC_Server_address + common_config.GrpcCustomerUI_RPC_Server_port //"127.0.0.1:50651"

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
// Simulate a www-Customer that Ask For Price

func requestPrice() {

	myEmptyParameter := &xxx.EmptyParameter{}
	resp, err := testClient.AskTaxiForPrice(context.Background(), myEmptyParameter)
	if err != nil {
		logMessagesWithError(4, "Could not send 'AskTaxiForPrice' to address: "+address_to_dial+". Error Message:", err)

	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'AskTaxiForPrice' on address "+address_to_dial+" successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)
		} else {
			logMessagesWithOutError(4, "'AskTaxiForPrice' on address "+address_to_dial+" NOT successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)

		}
	}

}

// ******************************************************************************
// Simulate a www-Customer that Accept Price

func accreptPrice() {

	myEmptyParameter := &xxx.EmptyParameter{}
	resp, err := testClient.AcceptPrice(context.Background(), myEmptyParameter)
	if err != nil {
		logMessagesWithError(4, "Could not send 'AcceptPrice' to address: "+address_to_dial+". Error Message:", err)

	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'AcceptPrice' on address "+address_to_dial+" successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)
		} else {
			logMessagesWithOutError(4, "'AcceptPrice' on address "+address_to_dial+" NOT successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)

		}
	}

}

// ******************************************************************************
// Simulate a www-Customer that Halts payments
func haltPayments(incomingRequest bool) {

	myhaltPaymentRequest := &xxx.HaltPaymentRequest{
		Haltpayment: incomingRequest,
	}

	resp, err := testClient.HaltPayments(context.Background(), myhaltPaymentRequest)
	if err != nil {
		logMessagesWithError(4, "Could not send 'haltPayments' to address: "+address_to_dial+". Error Message:", err)

	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'haltPayments' on address "+address_to_dial+" successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)
		} else {
			logMessagesWithOutError(4, "'haltPayments' on address "+address_to_dial+" NOT successfully processed")
			logMessagesWithOutError(4, "Response Message: "+resp.Comments)
			logMessagesWithOutError(4, "Payment Request: "+resp.Comments)

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
		log.Println("did not connect to Customer UI-Server on address: ", address_to_dial, "error message", err)
		os.Exit(0)
	} else {
		log.Println("gRPC connection OK to Customer UI- Server, address: ", address_to_dial)
		// Creates a new Clients
		testClient = xxx.NewCustomer_UIClient(remoteServerConnection)

	}

	fmt.Println("Enter 'requestPrice' to trigger get price")
	fmt.Println("Enter 'acceptPrice' to trigger accept of price")
	fmt.Println("Enter 'haltPayments(true)' to Halt Payments")
	fmt.Println("Enter 'haltPayments(false)' to Un-Halt Payments")
	fmt.Println("")

	for {

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		switch text {
		case "requestPrice":
			requestPrice()

			fmt.Println("Please try again!")
			fmt.Println("")

		case "acceptPrice":
			accreptPrice()

			fmt.Println("Please try again!")
			fmt.Println("")

		case "haltPayments(true)":
			haltPayments(true)

			fmt.Println("Please try again!")
			fmt.Println("")

		case "haltPayments(false)":
			haltPayments(false)

			fmt.Println("Please try again!")
			fmt.Println("")

		default:
			fmt.Println("Unknown command, please try again!")
			fmt.Println("")
		}
	}

}
