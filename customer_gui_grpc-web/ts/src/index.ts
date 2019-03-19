import {grpc} from "@improbable-eng/grpc-web";
import {Customer_UI} from "../../js_grpc_code/examplecom/library/customer_ui_api_pb_service";
import {EmptyParameter, AckNackResponse, PriceUnit, TimeUnit, HaltPaymentRequest, Price_UI, RPCMethods, UIPriceAndStateRespons} from "../../js_grpc_code/examplecom/library/customer_ui_api_pb";
///<reference path="typings/jquery/jquery.d.ts" />

declare const USE_TLS: boolean;
const host = USE_TLS ? "https://localhost:9091" : "http://localhost:9090";

//TODO Toggle Animation
// TODO


$( document ).ready(function() {
  prepareButtons();
});

// Prepare Buttons *****************'
function prepareButtons() {

  //AskTaxiForPrice-Button
  var asktaxiforpriceButton = <HTMLInputElement>document.getElementById("AskTaxiForPrice");
  asktaxiforpriceButton.addEventListener("click", Event => askTaxiForPrice());

  //AcceptPrice-Button
  var acceptpriceButton = <HTMLInputElement>document.getElementById("AcceptPrice");
  acceptpriceButton.addEventListener("click", Event => askTaxiForPrice());

  //HaltPayments-Button
  var haltpaymentsTrueButton = <HTMLInputElement>document.getElementById("HaltPayments");
  haltpaymentsTrueButton.addEventListener("click", Event => askTaxiForPrice());

  //UnHaltPayments-Button
  var haltpaymentsFalseButton = <HTMLInputElement>document.getElementById("UnHaltPayments");
  haltpaymentsFalseButton.addEventListener("click", Event => askTaxiForPrice());

  //LeaveTaxi-Button
  var leavetaxiButton = <HTMLInputElement>document.getElementById("LeaveTaxi");
  leavetaxiButton.addEventListener("click", Event => askTaxiForPrice());

}
//****************************

function controlButtonStates(allowedRPCMethods *client.RPCMethods) {

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



// Ask Taxi for Price. This is done to initiate connection with taxi. A message and Price is returned
function askTaxiForPrice() {
  const emptyParameter = new EmptyParameter()

  grpc.unary(Customer_UI.AskTaxiForPrice, {
    request: emptyParameter,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("askTaxiForPrice.onEnd.status", status, statusMessage);
      console.log("askTaxiForPrice.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("askTaxiForPrice.onEnd.message", message.toObject());
      }
      console.log("askTaxiForPrice.onEnd.trailers", trailers);
    }
  });
}

// Accept Price from Taxi. This is the second step in the process. AckNack is returned
function acceptPrice() {
  const emptyParameter = new EmptyParameter()

  grpc.unary(Customer_UI.AcceptPrice, {
    request: emptyParameter,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("acceptPrice.onEnd.status", status, statusMessage);
      console.log("acceptPrice.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("acceptPrice.onEnd.message", message.toObject());
      }
      console.log("acceptPrice.onEnd.trailers", trailers);
    }
  });
}


// Halt or UnHalt payments to Taxi. This is done during the ride by sending true/false AckNack is returned
function haltPayments(togglePaymentStream: boolean) {
  const haltPaymentRequest = new HaltPaymentRequest()
  haltPaymentRequest.setHaltpayment(togglePaymentStream)

  grpc.unary(Customer_UI.HaltPayments, {
    request: haltPaymentRequest,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("haltPayments.onEnd.status", status, statusMessage);
      console.log("haltPayments.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("haltPayments.onEnd.message", message.toObject());
      }
      console.log("haltPayments.onEnd.trailers", trailers);
    }
  });
}

// Leave Taxi. This is the last step in the process. AckNack is returned
function leaveTaxi() {
  const emptyParameter = new EmptyParameter()

  grpc.unary(Customer_UI.LeaveTaxi, {
    request: emptyParameter,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("leaveTaxi.onEnd.status", status, statusMessage);
      console.log("leaveTaxi.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("leaveTaxi.onEnd.message", message.toObject());
      }
      console.log("leaveTaxi.onEnd.trailers", trailers);
    }
  });
}

// UIPriceAndStateStream, Returns a stream with Price and State-info. This is started at beginning and continoues as long the system is connected to server
function receiveUIPriceAndStateStream() {

  // Create references to controlling buttons
  //var askTaxiForPriceButton = document.getElementById("AskTaxiForPrice");
  //var acceptPriceButton = document.getElementById("AcceptPrice");
  //var haltPaymentsButton = document.getElementById("HaltPayments");
  //var unHaltPaymentsButton = document.getElementById("UnHaltPayments");
  //var leaveTaxiButton = document.getElementById("LeaveTaxi");

  const emptyParameter = new EmptyParameter()

  const client = grpc.client(Customer_UI.UIPriceAndStateStream, {
    host: host,
  });

  client.onHeaders((headers: grpc.Metadata) => {
    console.log("UIPriceAndStateStream.onHeaders", headers);
  });
  client.onMessage((message: UIPriceAndStateRespons) => {
    console.log("UIPriceAndStateStream.onMessage", message.toObject());



    //Enable or Disable Buttons depending on state stream from server
    var myMap;
    myMap = message.toObject().allowedrpcmethods;
    var asktaxiforprice = JSON.parse(JSON.stringify(myMap)).asktaxiforprice;
    var acceptprice = JSON.parse(JSON.stringify(myMap)).acceptprice;
    var haltpaymentsTrue = JSON.parse(JSON.stringify(myMap)).haltpaymentsTrue;
    var haltpaymentsFalse = JSON.parse(JSON.stringify(myMap)).haltpaymentsFalse;
    var leavetaxi = JSON.parse(JSON.stringify(myMap)).leavetaxi;

    //AskTaxiForPrice-Button
    var asktaxiforpriceButton = <HTMLInputElement> document.getElementById("AskTaxiForPrice");
    asktaxiforpriceButton.disabled = !asktaxiforprice;


    //AcceptPrice-Button
    var acceptpriceButton = <HTMLInputElement> document.getElementById("AcceptPrice");
    acceptpriceButton.disabled = !acceptprice;

    //HaltPayments-Button
    var haltpaymentsTrueButton = <HTMLInputElement> document.getElementById("HaltPayments");
    haltpaymentsTrueButton.disabled = !haltpaymentsTrue;

    //UnHaltPayments-Button
    var haltpaymentsFalseButton = <HTMLInputElement> document.getElementById("UnHaltPayments");
    haltpaymentsFalseButton.disabled = !haltpaymentsFalse;

    //LeaveTaxi-Button
    var leavetaxiButton = <HTMLInputElement> document.getElementById("LeaveTaxi");
    leavetaxiButton.disabled = !leavetaxi;

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


  });
  client.onEnd((code: grpc.Code, msg: string, trailers: grpc.Metadata) => {
    console.log("UIPriceAndStateStream.onEnd", code, msg, trailers);
  });
  client.start();
  client.send(emptyParameter);
}


/*
function getBook() {
  const getBookRequest = new GetBookRequest();
  getBookRequest.setIsbn(60929871);
  grpc.unary(BookService.GetBook, {
    request: getBookRequest,
    host: host,
    onEnd: res => {
      const { status, statusMessage, headers, message, trailers } = res;
      console.log("getBook.onEnd.status", status, statusMessage);
      console.log("getBook.onEnd.headers", headers);
      if (status === grpc.Code.OK && message) {
        console.log("getBook.onEnd.message", message.toObject());
      }
      console.log("getBook.onEnd.trailers", trailers);
      queryBooks();
    }
  });
}
*/

//getBook();
//askTaxiForPrice();
//acceptPrice()
receiveUIPriceAndStateStream();
//prepareButtons()



/*

function queryBooks() {
  const queryBooksRequest = new QueryBooksRequest();
  queryBooksRequest.setAuthorPrefix("Geor");
  const client = grpc.client(BookService.QueryBooks, {
    host: host,
  });
  client.onHeaders((headers: grpc.Metadata) => {
    console.log("queryBooks.onHeaders", headers);
  });
  client.onMessage((message: Book) => {
    console.log("queryBooks.onMessage", message.toObject());
  });
  client.onEnd((code: grpc.Code, msg: string, trailers: grpc.Metadata) => {
    console.log("queryBooks.onEnd", code, msg, trailers);
  });
  client.start();
  client.send(queryBooksRequest);
}

*/
