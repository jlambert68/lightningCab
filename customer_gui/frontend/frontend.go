package main

import (
	"strings"

	"honnef.co/go/js/dom"

	"jlambert/lightningCab/grpc_api/proto/client"
	"github.com/gopherjs/gopherjs/js"

	"context"
	"io"
	"strconv"
)

// Build this snippet with GopherJS, minimize the output and
// write it to html/frontend.js.
//go:generate gopherjs build frontend.go -m -o html/frontend.js

// Zopfli compress static files.
//go:generate find ./html/ -name *.gz -prune -o -type f -exec go-zopfli {} +

// Integrate generated JS into a Go file for static loading.
//go:generate bash -c "go run assets_generate.go"

// This constant is very useful for interacting with
// the DOM dynamically
var document = dom.GetWindow().Document().(dom.HTMLDocument)

// Define no-op main since it doesn't run when we want it to
func main() {}

// Ensure our setup() gets called as soon as the DOM has loaded
func init() {
	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		go setup()
	})
}

// Setup is where we do the real work.
func setup() {
	// This is the address to the server, and should be used
	// when creating clients.
	serverAddr := strings.TrimSuffix(document.BaseURI(), "/")

	// TODO: Use functions exposed by generated interface
	customerBackendServerClient := client.NewCustomer_UIClient(serverAddr)
	//cli_stream := client.NewCustomerUIPriceStreamClient(serverAddr)

	//document.Body().SetInnerHTML(`<div><h2>GopherJS gRPC-Web is great! JL</h2></div>`)

	//***********************************
	// Functions for Buttons

	d := dom.GetWindow().Document()

	// Button Ask Taxi For Price
	askTaxiForPrice := d.GetElementByID("AskTaxiForPrice").(*dom.HTMLButtonElement)
	askTaxiForPrice.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed  'Ask Taxi For Price'-Button")
		go func() {
			resp, err := customerBackendServerClient.AskTaxiForPrice(context.Background(), &client.EmptyParameter{})
			if err != nil {
				print("Error in respons from 'Accept Price'")
			}
			print(resp)
		}()
		//js.Global.Call("toggleAnimation", true)

	})

	// Button Accept Price
	acceptPrice := d.GetElementByID("AcceptPrice").(*dom.HTMLButtonElement)
	acceptPrice.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Accept Price'-Button")
		go func() {
			resp, err := customerBackendServerClient.AcceptPrice(context.Background(), &client.EmptyParameter{})
			if err != nil {
				print("Error in respons from 'Accept Price'")
			}
			print(resp)
		}()

		//js.Global.Call("toggleAnimation", true)

	})

	// Button Halt Payments
	haltPayments := d.GetElementByID("HaltPayments").(*dom.HTMLButtonElement)
	haltPayments.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Halt'-Button")

		js.Global.Call("toggleAnimation", true)
		go func() {
			resp, err := customerBackendServerClient.HaltPayments(context.Background(), &client.HaltPaymentRequest{Haltpayment: true})
			if err != nil {
				print("Error in respons from 'Accept Price'")
			}
			print(resp)
		}()
	})

	// Button Un-Halt Payments
	unHaltPayments := d.GetElementByID("UnHaltPayments").(*dom.HTMLButtonElement)
	unHaltPayments.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Un-Halt'-Button")

		js.Global.Call("toggleAnimation", false)
		go func() {
			resp, err := customerBackendServerClient.HaltPayments(context.Background(), &client.HaltPaymentRequest{Haltpayment: false})
			if err != nil {
				print("Error in respons from 'Accept Price'")
			}
			print(resp)
		}()
	})

	// Button LeaveTaxi
	leaveTaxi := d.GetElementByID("LeaveTaxi").(*dom.HTMLButtonElement)
	leaveTaxi.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed Halt-Button")
		go func() {
			resp, err := customerBackendServerClient.LeaveTaxi(context.Background(), &client.EmptyParameter{})
			if err != nil {
				print("Error in respons from 'Accept Price'")
			}
			print(resp)
		}()
		//js.Global.Call("toggleAnimation", true)

	})

	// Receive Payment Info and State Info for buttons
	// Wrapped in goroutine because Recv is blocking
	go func() {
		print("Entering function for receiving 'paymentDataAndStateInfo'")
		//ctx, cancel := context.WithCancel(context.Background())
		//defer cancel()
		//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		//defer cancel()
		ctx := context.Background()

		srv, err := customerBackendServerClient.UIPriceAndStateStream(ctx, &client.EmptyParameter{})
		if err != nil {
			println(err)
			return
		} else {
			print("Success in connecting to 'UIPriceAndStateStream'")

		}

		for {
			// Blocks until response is received
			paymentDataAndStateInfo, err := srv.Recv()

			if err == io.EOF {
				print("HMMM, OF for Stream, skumt borde inte slutat här när vi tar emot paymentDataAndStateInfo-Stream, borde avsluta Customer")
				print(paymentDataAndStateInfo)
				print(err)
				break
			}
			if err != nil {
				print("Problem when streaming from Customer Stream:")
				print(paymentDataAndStateInfo)
				print(err)
				break
			}

			//print(paymentDataAndStateInfo)
			// Set States for Buttons
			controlButtonStates(paymentDataAndStateInfo.AllowedRPCMethods)

			// Set Wallet balance
			setWalletbalance(int64(paymentDataAndStateInfo.CurrentWalletbalanceSatoshi), float32(paymentDataAndStateInfo.CurrentWalletbalanceSek))

			// Set Taxi Ride cost
			setTaxiRideCost(int64(paymentDataAndStateInfo.CurrentTaxirideSatoshi), float32(paymentDataAndStateInfo.CurrentTaxiRideSek))

			//Set Average Payment Amount and No of Payments
			setAveragePaymentAmount(
				int64(paymentDataAndStateInfo.AveragePaymentAmountSatoshi),
				float32(paymentDataAndStateInfo.AveragePaymentAmountSek),
				float32(paymentDataAndStateInfo.AvaregeNumberOfPayments))

			// Depending of Accelation, Speed and time(paymentintervall) the cats dimensions and rotation speed will change
			setCatDimensions(int32(paymentDataAndStateInfo.AccelarationPercent), int32(paymentDataAndStateInfo.SpeedPercent), float32(paymentDataAndStateInfo.AvaregeNumberOfPayments))

		}
	}()

}

func controlButtonStates(allowedRPCMethods *client.RPCMethods) {

	//Get Button-objects
	d := dom.GetWindow().Document()
	askTaxiForPriceButton := d.GetElementByID("AskTaxiForPrice").(*dom.HTMLButtonElement)
	acceptPriceButton := d.GetElementByID("AcceptPrice").(*dom.HTMLButtonElement)
	haltPaymentsButton := d.GetElementByID("HaltPayments").(*dom.HTMLButtonElement)
	unHaltPaymentsButton := d.GetElementByID("UnHaltPayments").(*dom.HTMLButtonElement)
	leaveTaxiButton := d.GetElementByID("LeaveTaxi").(*dom.HTMLButtonElement)

	// Enable and Disable Buttons
	askTaxiForPriceButton.Disabled = !allowedRPCMethods.AskTaxiForPrice
	acceptPriceButton.Disabled = !allowedRPCMethods.AcceptPrice
	haltPaymentsButton.Disabled = !allowedRPCMethods.HaltPaymentsTrue
	unHaltPaymentsButton.Disabled = !allowedRPCMethods.HaltPaymentsFalse
	leaveTaxiButton.Disabled = !allowedRPCMethods.LeaveTaxi


}

func setWalletbalance(walletAmountSatoshi int64, walletAmountSEK float32) {

	// Get text objecta for wallet balance
	d := dom.GetWindow().Document()
	walletAmountSatoshiText := d.GetElementByID("walletAmountSatoshi").(*dom.HTMLTableCellElement)
	walletAmountSEKText := d.GetElementByID("walletAmountSEK").(*dom.HTMLTableCellElement)

	walletAmountSatoshiTextValue := strconv.FormatInt(walletAmountSatoshi,10)
	walletAmountSEKTextValue := strconv.FormatFloat(float64(walletAmountSEK), 'f', 6, 64)

	walletAmountSatoshiText.SetInnerHTML(walletAmountSatoshiTextValue)
	walletAmountSEKText.SetInnerHTML(walletAmountSEKTextValue)

	}

func setTaxiRideCost(taxiAmountSatoshi int64, taxiAmountSEK float32) {

	// Get text objects for taxi cost
	d := dom.GetWindow().Document()
	taxiAmountSatoshiText := d.GetElementByID("taxiRideAmountSatoshi").(*dom.HTMLTableCellElement)
	taxiAmountSEKText := d.GetElementByID("taxiRideAmountSEK").(*dom.HTMLTableCellElement)

	taxiAmountSatoshiTextValue := strconv.FormatInt(taxiAmountSatoshi,10)
	taxiAmountSEKTextValue := strconv.FormatFloat(float64(taxiAmountSEK), 'f', 6, 64)

	taxiAmountSatoshiText.SetInnerHTML(taxiAmountSatoshiTextValue)
	taxiAmountSEKText.SetInnerHTML(taxiAmountSEKTextValue)
}

func setAveragePaymentAmount(averageSatoshi int64, averageSEK float32, averageNoOfPayments float32) {

	// Get text objects for taxi cost
	d := dom.GetWindow().Document()
	averageSatoshiText := d.GetElementByID("averageAmountSatoshi").(*dom.HTMLTableCellElement)
	averageSEKText := d.GetElementByID("averageAmountSEK").(*dom.HTMLTableCellElement)
	averageNoPaymentsText := d.GetElementByID("averageNoPayments").(*dom.HTMLTableCellElement)

	averageSatoshiTextValue := strconv.FormatInt(averageSatoshi,10)
	averageSEKTextValue := strconv.FormatFloat(float64(averageSEK), 'f', 6, 64)
	averageNoPaymentsTextValue := strconv.FormatFloat(float64(averageNoOfPayments), 'f', 6, 64)

	averageSatoshiText.SetInnerHTML(averageSatoshiTextValue)
	averageSEKText.SetInnerHTML(averageSEKTextValue)
	averageNoPaymentsText.SetInnerHTML(averageNoPaymentsTextValue)

}

func setCatDimensions(accelerationPercent int32, acceleationSpeed int32, averagePaymentPerSecond float32) {
	js.Global.Call("changeCatProperties", accelerationPercent, acceleationSpeed, averagePaymentPerSecond)
}
