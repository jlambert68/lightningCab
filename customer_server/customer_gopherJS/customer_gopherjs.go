package main

import "github.com/gopherjs/jquery"
import "honnef.co/go/js/dom"
import "github.com/gopherjs/gopherjs/js"

//convenience:
var jQuery = jquery.NewJQuery

//const (
//INPUT  = "input#name
//OUTPUT = "span#output"
//HALTBUTTON = "button#HaltPayments"
//)

func main() {

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)
	d := dom.GetWindow().Document()

	// Button Ask Taxi For Price
	askTaxiForPrice := d.GetElementByID("AskTaxiForPrice").(*dom.HTMLButtonElement)
	askTaxiForPrice.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed  'Ask Taxi For Price'-Button")

		//js.Global.Call("toggleAnimation", true)


	})

	// Button Accept Price
	acceptPrice := d.GetElementByID("AcceptPrice").(*dom.HTMLButtonElement)
	acceptPrice.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Accept Price'-Button")

		//js.Global.Call("toggleAnimation", true)

	})

	// Button Halt Payments
	haltPayments := d.GetElementByID("HaltPayments").(*dom.HTMLButtonElement)
	haltPayments.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Halt'-Button")

		js.Global.Call("toggleAnimation", true)

	})

	// Button Un-Halt Payments
	unHaltPayments := d.GetElementByID("UnHaltPayments").(*dom.HTMLButtonElement)
	unHaltPayments.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed 'Un-Halt'-Button")

		js.Global.Call("toggleAnimation", false)

	})

	// Button LeaveTaxi
	leaveTaxi := d.GetElementByID("LeaveTaxi").(*dom.HTMLButtonElement)
	leaveTaxi.AddEventListener("click", false, func(event dom.Event) {
		print("User pressed Halt-Button")

		//js.Global.Call("toggleAnimation", true)


	})

}
