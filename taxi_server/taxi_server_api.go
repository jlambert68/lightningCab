package main

import (
	"log"
	"fmt"
	"time"
	//"google.golang.org/grpc"
	"golang.org/x/net/context"
	taxi_api "github.com/jlambert68/lightningCab/grpc_api/taxi_grpc_api"
	"github.com/jlambert68/lightningCab/common_config"
	"math"
	"github.com/sirupsen/logrus"
)

/*
rpc AskTaxiForPrice (Enviroment) returns (Price) {
}

//Accept price from Taxi
rpc AcceptPrice (Enviroment) returns (AckNackResponse) {
}

// Returns a stream of PaymentRequests to Customer
rpc PaymentRequestStream (Enviroment) returns (stream PaymentRequest) {
}
*/

var averageTimeBetweenIncoicesToCustomer []int64

// Customer connects and ask for price
func (s *taxiServiceServer) AskTaxiForPrice(ctx context.Context, environment *taxi_api.Enviroment) (*taxi_api.Price, error) {

	//log.Println("Incoming: 'AskTaxiForPrice'")
	taxi.logger.Info("Incoming: 'AskTaxiForPrice'")
	//fmt.Println("sleeping...for 3 seconds")
	taxi.logger.Info("sleeping...for 3 seconds")
	time.Sleep(3 * time.Second)

	currentPrice := &taxi_api.Price{
		Acknack: false,
		Comments: "",
		Speed: int64(math.Round(float64(common_config.SpeedSatoshiPerSecond))),
		Acceleration: int64(math.Round(float64(common_config.MaxAccelarationSatoshiPerSecond))),
		Time: int64(math.Round(float64(common_config.TimeSatoshiPerSecond))),
		Timeunit: taxi_api.TimeUnit_SecondsBetweenPaymentmentRequests,
		PaymentRequestInterval: 1,
		Priceunit: taxi_api.PriceUnit_SatoshiPerSecond}

	// Check if State machine accepts State change
	err := taxi.CustomerConnects(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case taxi_api.TestOrProdEnviroment_Test:
			// Customer Connects to Taxi
			//log.Println("Customer Connects to Taxi:")
			taxi.logger.Info("Customer Connects to Taxi:")
			currentPrice.Acknack = true
			currentPrice.Comments = "Welcome to the best Taxi you can find around here!"

		case taxi_api.TestOrProdEnviroment_Production:
			// Customer Connects to Taxi
			//log.Println("Customer Connects to Taxi:")
			taxi.logger.Info("Customer Connects to Taxi:")
			currentPrice.Acknack = true
			currentPrice.Comments = "Welcome to the best Taxi you can find around here!"

		default:
			logMessagesWithOutError(4, "Unknown incomming enviroment: "+environment.TestOrProduction.String())
			currentPrice.Acknack = false
			currentPrice.Comments = "Unknown incomming enviroment: " + environment.TestOrProduction.String()
		}

		err = taxi.CustomerConnects(false)
		if err != nil {
			logMessagesWithError(4, "There was a problem to change state in Taxi State machine: ", err)
			currentPrice.Acknack = false
			currentPrice.Comments = "There was a problem to change state in Taxi State machine"
		}

	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able have customer connects to it: ", err)
		currentPrice.Acknack = false
		currentPrice.Comments = "State machine is not in correct state to be able have customer connects to it"
	}

	return currentPrice, nil

}

// Customer accepts price
func (s *taxiServiceServer) AcceptPrice(ctx context.Context, environment *taxi_api.Enviroment) (*taxi_api.AckNackResponse, error) {

	log.Println("Incoming: 'AcceptPrice'")
	fmt.Println("sleeping...for 3 seconds")
	time.Sleep(3 * time.Second)

	var acknack bool
	var returnMessage string

	// Check if State machine accepts State change
	err := taxi.CustomerAcceptsPrice(true)

	if err == nil {

		// Check if to Simulate or not
		switch environment.TestOrProduction {

		case taxi_api.TestOrProdEnviroment_Test:
			// Customer Connects to Taxi
			//log.Println("Customer Connects to Taxi:")
			taxi.logger.Info("Customer Connects to Taxi:")
			acknack = true
			returnMessage = "Welcome to the best Taxi you can find around here!"

		case taxi_api.TestOrProdEnviroment_Production:
			// Customer Connects to Taxi
			//log.Println("Customer Connects to Taxi:")
			taxi.logger.Info("Customer Connects to Taxi:")
			acknack = true
			returnMessage = "Welcome to the best Taxi you can find around here!"

		default:
			logMessagesWithOutError(4, "Unknown incomming enviroment: "+environment.TestOrProduction.String())
			acknack = false
			returnMessage = "Unknown incomming enviroment: " + environment.TestOrProduction.String()
		}

		err = taxi.CustomerAcceptsPrice(false)
		if err != nil {
			logMessagesWithError(4, "There was a problem to change state in Taxi State machine: ", err)
			acknack = false
			returnMessage = "There was a problem to change state in Taxi State machine"
		}

	} else {

		logMessagesWithError(4, "State machine is not in correct state to be able have customer connects to it: ", err)
		acknack = false
		returnMessage = "State machine is not in correct state to be able have customer connects to it"
	}

	return &taxi_api.AckNackResponse{Acknack: acknack, Comments: returnMessage}, nil
}

func (s *taxiServiceServer) PaymentRequestStream(enviroment *taxi_api.Enviroment, stream taxi_api.Taxi_PaymentRequestStreamServer) (err error) {
	//log.Printf("Incoming: 'PaymentRequestStream'")
	taxi.logger.Info("Incoming: 'PaymentRequestStream'")
	var firstTime bool = true
	//	var firstMissedPaymentTimer time.Timer
	//	var lastMissedPaymentTimer time.Timer

	abortPaymentRequestGeneration = false

	err = nil
	paymentRequestResponse := &taxi_api.PaymentRequest{
		LightningPaymentRequest: "",
		SpeedAmountSatoshi: 0,
		AccelerationAmountSatoshi: 0,
		TimeAmountSatoshi: 0,
		SpeedAmountSek: 0,
		AccelerationAmountSek: 0,
		TimeAmountSek: 0,
		TotalAmountSatoshi: 0,
		TotalAmountSek: 0,
		AccelerationPercent:0,
		SpeedPercent:0,
	}


	firstMissedPaymentTimer := time.NewTimer(common_config.SecondsBeforeFirstPaymentTimeOut * time.Second)
	lastMissedPaymentTimer := time.NewTimer(common_config.SecondsBeforeSecondPaymentTimeOut * time.Second)

	for {
		/*// Stop waiting for paid payment requests
		if abortPaymentRequestGeneration == true {
			break
		}*/


		if paymentRequestIsPaid == true || firstTime == true {
			// Let Engine have power
			taxi.CutPowerToEngine(false)

			if firstTime == true {
				//log.Printf("First time creating PaymentRequest")
				taxi.logger.Info("First time creating PaymentRequest")
			} else {
				//log.Printf("Not first time creating PaymentRequest")
				taxi.logger.Info("Not first time creating PaymentRequest")
				taxi.invoiceStatistics.invoicesPaid = taxi.invoiceStatistics.invoicesPaid + 1
			}


			firstTime = false
			firstMissedPaymentTimeOut = false
			lastMissedPaymentTimeOut = false

			// Populate stream Respons
			invoicedataToCustomer, err := generateInvoice()

			//spew.Println(invoicedataToCustomer)
			//log.Println("Invoice Received from LND-generator: " + invoicedataToCustomer.Invoice)

			paymentRequestResponse.LightningPaymentRequest = invoicedataToCustomer.Invoice
			paymentRequestResponse.AccelerationAmountSatoshi = lastPaymentData.lastReceivedAmountdata.accelerationAmount
			paymentRequestResponse.SpeedAmountSatoshi = lastPaymentData.lastReceivedAmountdata.speedAmount
			paymentRequestResponse.TimeAmountSatoshi = lastPaymentData.lastReceivedAmountdata.timeAmount
			paymentRequestResponse.AccelerationAmountSek = float32(lastPaymentData.lastReceivedAmountdata.accelerationAmount) * common_config.BTCSEK / common_config.SatoshisPerBTC
			paymentRequestResponse.SpeedAmountSek = float32(lastPaymentData.lastReceivedAmountdata.speedAmount) * common_config.BTCSEK / common_config.SatoshisPerBTC
			paymentRequestResponse.TimeAmountSek = float32(lastPaymentData.lastReceivedAmountdata.timeAmount) * common_config.BTCSEK / common_config.SatoshisPerBTC
			paymentRequestResponse.TotalAmountSatoshi = lastPaymentData.lastAmountToPay_satoshi
			paymentRequestResponse.TotalAmountSek = lastPaymentData.lastAmountToPay_sek
			paymentRequestResponse.AccelerationPercent = lastPaymentData.lastPowerMessaurment.Acceleration
			paymentRequestResponse.SpeedPercent = lastPaymentData.lastPowerMessaurment.Speed

			taxi.logger.Debug("invoicedataToCustomer:", invoicedataToCustomer)
			//spew.Println(paymentRequestResponse)

			if err == nil {
				paymentRequestIsPaid = false

				if err := stream.Send(paymentRequestResponse); err != nil {
					//log.Printf("Error when streaming back: 'PaymentRequestStream'")
					taxi.logger.Error("Error when streaming back: 'PaymentRequestStream", err)
					return err
					break
				} else {
					//log.Println("Invoice sent to customer: " + paymentRequestResponse.LightningPaymentRequest)
					taxi.logger.WithFields(logrus.Fields{
						"paymentRequestResponse.LightningPaymentRequest":    paymentRequestResponse.LightningPaymentRequest,
					}).Info("Invoice sent to customer:")
					taxi.invoiceStatistics.invoicedCreated = taxi.invoiceStatistics.invoicedCreated + 1
				}
			} else {
				// TotalAmountSatoshi == 0
				//log.Printf("Error, can't stream due to 'paymentRequestResponse.TotalAmountSatoshi == 0'")
				logMessagesWithError(4, "Error, can't stream due to 'paymentRequestResponse.TotalAmountSatoshi == 0'", err)
			}


			/*			if firstTime == true {
							firstMissedPaymentTimer := time.NewTimer(common_config.SecondsBeforeFirstPaymentTimeOut * time.Second)
							lastMissedPaymentTimer := time.NewTimer(common_config.SecondsBeforeSecondPaymentTimeOut * time.Second)
						} else {*/
			firstMissedPaymentTimer.Reset(common_config.SecondsBeforeFirstPaymentTimeOut * time.Second)
			lastMissedPaymentTimer.Reset(common_config.SecondsBeforeSecondPaymentTimeOut * time.Second)
			//}

			go func() {
				<-firstMissedPaymentTimer.C
				//log.Println("Timer 'firstMissedPaymentTimer' expired")
				taxi.logger.Warning("Timer 'firstMissedPaymentTimer' expired")
				firstMissedPaymentTimeOut = true

			}()
			go func() {
				<-lastMissedPaymentTimer.C
				//log.Println("Timer 'lastMissedPaymentTimer' expired")
				taxi.logger.Warning("Timer 'lastMissedPaymentTimer' expired")
				lastMissedPaymentTimeOut = true
				abortPaymentRequestGeneration = true
			}()

		} /*else {
			log.Println("paymentRequestIsPaid =", paymentRequestIsPaid)
			log.Println("firstTime =", firstTime)
		}*/

		averageTime := calculateAverageTimeBetweenInvoices()
		if averageTime > common_config.MilliSecondsBetweenPaymentRequest {
			timeToSleep := time.Duration(common_config.MilliSecondsBetweenPaymentRequest * (common_config.MilliSecondsBetweenPaymentRequest / averageTime))
			time.Sleep( timeToSleep * time.Millisecond)
		} else {
			time.Sleep(common_config.MilliSecondsBetweenPaymentRequest * time.Millisecond)
		}


		// First Timeout
		if paymentRequestIsPaid == false && firstMissedPaymentTimeOut == true {
			//Check if posible to change state
			if taxi.PaymentsStopsComing(true) == nil {
				//log.Printf("PaymentRequest not paid in time. Stop Taxi or continue stop and wait for payment")
				taxi.logger.Warning("PaymentRequest not paid in time. Stop Taxi or continue stop and wait for payment")
				taxi.PaymentsStopsComing(false)
				taxi.CutPowerToEngine(true)
			} else {
				//log.Println("This should not happend for firstTimer")
				taxi.logger.Error("This should not happend for firstTimer")
				//log.Println("Current State", taxi.TaxiStateMachine.State())
				taxi.logger.WithFields(logrus.Fields{
					"taxi.TaxiStateMachine.State()":    taxi.TaxiStateMachine.State(),
				}).Error("Current State:")
				//lastMissedPaymentTimer.Stop()
				break
			}
			firstMissedPaymentTimeOut = false
		}

		// Second and Last Timeout
		if paymentRequestIsPaid == false && lastMissedPaymentTimeOut == true {
			//log.Printf("PaymentRequest timedout. Get Taxi ready for new customer")
			taxi.logger.Warning("PaymentRequest timedout. Get Taxi ready for new customer")
			//Stop Stream for a while and send state machine in wait mode

			//Check if posible to change state
			if taxi.abortPaymentRequestGeneration(true) == nil {
				//log.Println("Abort Payment Request Generation and make Taxi ready for next customer")
				taxi.logger.Warning("Abort Payment Request Generation and make Taxi ready for next customer")
				taxi.abortPaymentRequestGeneration(false)
				taxi.CutPowerToEngine(true)
			} else {
				//log.Println("This should not happend for lastTimer")
				taxi.logger.Error("This should not happend for lastTimer")
				//log.Println("Current State", taxi.TaxiStateMachine.State())
				taxi.logger.WithFields(logrus.Fields{
					"taxi.TaxiStateMachine.State()":    taxi.TaxiStateMachine.State(),
				}).Error("Current State:")
				break
			}
			//log.Println("From LastTimer Existing Generate PaymentRequests")
			taxi.logger.Info("From LastTimer Existing Generate PaymentRequests")
			break
		}

	}

	//log.Println("Leaving 'func PaymentRequestStream'")
	taxi.logger.Info("Leaving 'func PaymentRequestStream'")
	return nil
}

// Calculate Average time between invoices to customer
func calculateAverageTimeBetweenInvoices() (int64) {


	// Get current Unix time in seconds
	now := time.Now()
	secs := now.Unix()


	if len(averageTimeBetweenIncoicesToCustomer) != 0 {
		for {
			// Remove all data objects that has an old timestamp
			if secs-averageTimeBetweenIncoicesToCustomer[0] > common_config.TimeForAveragePaymentCalculation {
				// Subract old data from latest and Remove post due to old data

				averageTimeBetweenIncoicesToCustomer = averageTimeBetweenIncoicesToCustomer[1:]
				if len(averageTimeBetweenIncoicesToCustomer) == 0 {
					break
				}
			} else {

				break
			}
		}
	}

	// Add the new payment object
	averageTimeBetweenIncoicesToCustomer = append(averageTimeBetweenIncoicesToCustomer, secs)

	averageTime := int64(1000 / (float32(len(averageTimeBetweenIncoicesToCustomer)) / common_config.TimeForAveragePaymentCalculation))
	return averageTime
}

