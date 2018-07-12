package main

import (
	"strings"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"log"
	"fmt"
	"os"
	//taxoTotollGate_api "jlambert/lightningCab/toll_road_server/toll_road_grpc_api"
	"jlambert/lightningCab/common_config"
	"bufio"

	"jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"strconv"
	"time"
	"io"
	"jlambert/lightningCab/customer_server/lightningConnection"
)

var (
	// Standard Taxi gRPC Server
	remoteTaxiServerConnection *grpc.ClientConn
	customerClient             taxi_grpc_api.TaxiClient

	taxi_address_to_dial string                    = common_config.GrpcTaxiServer_address + common_config.GrpcTaxiServer_port
	useEnv                                         = taxi_grpc_api.TestOrProdEnviroment_Test
	useEnvironment       *taxi_grpc_api.Enviroment = &taxi_grpc_api.Enviroment{TestOrProduction: useEnv}
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
// Simulate a Customer that ask for a price
func askTaxiForPrice() {

	resp, err := customerClient.AskTaxiForPrice(context.Background(), useEnvironment)
	if err != nil {
		logMessagesWithError(4, "Could not send 'AskTaxiForPrice' to address: "+taxi_address_to_dial+". Error Message:", err)

	} else {

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

		}
	}

}

// ******************************************************************************
// Simulate a Customer that accepts the price
func acceptPrice() {

	resp, err := customerClient.AcceptPrice(context.Background(), useEnvironment)
	if err != nil {
		logMessagesWithError(4, "Could not send 'AcceptPrice' to address: "+taxi_address_to_dial+". Error Message:", err)

	} else {

		if resp.GetAcknack() == true {
			logMessagesWithOutError(4, "'AcceptPrice' on address "+taxi_address_to_dial+" successfully processed")
			logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)

			go receiveTaxiInvoices(customerClient, useEnvironment)

		} else {
			logMessagesWithOutError(4, "'AcceptPrice' on address "+taxi_address_to_dial+" NOT successfully processed")
			logMessagesWithOutError(4, "Response Message (Comments): "+resp.Comments)

		}
	}
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
	for {
		invoice, err := stream.Recv()
		if err == io.EOF {
			log.Println("HMMM, skumt borde inte slutat här när vi tar emot InvoiceStream, borde avsluta Customer")
			break
		}
		if err != nil {
			log.Fatalf("Problem when streaming from Taxi invoice Stream:", client, err)
		}
		//Customer Pays Invoice
		invoices[0] = invoice.LightningPaymentRequest
		err = lightningConnection.PayReceivedInvoicesFromTaxi(invoices)
		if err != nil {
			log.Println("Problem when paying Invoice from Taxi: ", err)
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

		remoteTaxiServerConnection.Close()
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

	fmt.Println("Enter 'askTaxiForPrice' to connect to Taxi")
	fmt.Println("Enter 'acceptPrice' to accept price and start incoming invoice stream")
	fmt.Println("")

	for {

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		switch text {
		case "askTaxiForPrice":
			askTaxiForPrice()

			fmt.Println("Please try again!")
			fmt.Println("")

		case "acceptPrice":
			acceptPrice()

			fmt.Println("Please try again!")
			fmt.Println("")

		default:
			fmt.Println("Unknown command, please try again!")
			fmt.Println("")
		}
	}

}
