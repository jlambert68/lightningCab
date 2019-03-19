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
// Enable or Disable buttons depending on state from Server
function controlButtonStates(allowedRPCMethods:any) {


  var asktaxiforprice = JSON.parse(JSON.stringify(allowedRPCMethods)).asktaxiforprice;
  var acceptprice = JSON.parse(JSON.stringify(allowedRPCMethods)).acceptprice;
  var haltpaymentsTrue = JSON.parse(JSON.stringify(allowedRPCMethods)).haltpaymentsTrue;
  var haltpaymentsFalse = JSON.parse(JSON.stringify(allowedRPCMethods)).haltpaymentsFalse;
  var leavetaxi = JSON.parse(JSON.stringify(allowedRPCMethods)).leavetaxi;

  //Get Button-objects
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


}

function setWalletbalance(walletAmountSatoshi:number, walletAmountSEK:number) {

  // Get text objects for wallet balance
  var walletAmountSatoshiText = <HTMLTableCellElement> document.getElementById("walletAmountSatoshi");
  var walletAmountSEKText = <HTMLTableCellElement> document.getElementById("walletAmountSEK");

  // Convert values into text
  var walletAmountSatoshiTextValue = walletAmountSatoshi.toString();
  var walletAmountSEKTextValue = walletAmountSEK.toFixed(6);

  // Set text-values
  walletAmountSatoshiText.innerHTML = walletAmountSatoshiTextValue;
  walletAmountSEKText.innerHTML = walletAmountSEKTextValue;

}

function setTaxiRideCost(taxiAmountSatoshi:number, taxiAmountSEK:number) {

  // Get text objects for taxi cost
  var taxiAmountSatoshiText = <HTMLTableCellElement> document.getElementById("taxiRideAmountSatoshi");
  var taxiAmountSEKText = <HTMLTableCellElement> document.getElementById("taxiRideAmountSEK");

  // Convert values into text
  var taxiAmountSatoshiTextValue = taxiAmountSatoshi.toString();
  var taxiAmountSEKTextValue = taxiAmountSEK.toFixed(6);

  // Set text-values
  taxiAmountSatoshiText.innerHTML = taxiAmountSatoshiTextValue;
  taxiAmountSEKText.innerHTML = taxiAmountSEKTextValue;

}

function setAveragePaymentAmount(averageSatoshi:number, averageSEK:number, averageNoOfPayments:number) {

  // Get text objects for taxi cost
  var averageSatoshiText = <HTMLTableCellElement> document.getElementById("averageAmountSatoshi");
  var averageSEKText = <HTMLTableCellElement> document.getElementById("averageAmountSEK");
  var averageNoPaymentsText = <HTMLTableCellElement> document.getElementById("averageNoPayments");

  // Convert values into text
  var averageSatoshiTextValue = averageSatoshi.toString();
  var averageSEKTextValue = averageSEK.toFixed(6);
  var averageNoPaymentsTextValue = averageNoOfPayments.toFixed(6);

  // Set text-values
  averageSatoshiText.innerHTML = averageSatoshiTextValue;
  averageSEKText.innerHTML = averageSEKTextValue;
  averageNoPaymentsText.innerHTML = averageNoPaymentsTextValue;

}


function setCatDimensions(acceleation:number, speed:number, rotation:number) {

  var style;
  var catWidth;
  var catHeigt;
  var catTopMarign;
  var catRightMargin;
  var catBottomMargin;
  var catLeftMargin;
  var catOriginalWidth;
  var catOriginalHeight;

  catOriginalWidth = 120;
  catOriginalHeight = 120
  catWidth = acceleation / 100 * 3 * 120 + 120;
  catHeigt = speed / 100 * 3 * 120 + 120;

  catTopMarign = - catHeigt / 2;
  catRightMargin = 0;
  catBottomMargin = 0;
  catLeftMargin = - catWidth / 2;

  var catImage = <HTMLImageElement> document.getElementById("catImageId");
  style = catImage.style

  style.width = catWidth.toString() + "px";
  style.height = catHeigt.toString() + "px";
  style.marginTop = catTopMarign.toString() + "px";
  //style.marginright = catRightMargin.toString() + "px";
  //style.marginbottom = catBottomMargin.toString() + "px";
  style.marginLeft = catLeftMargin.toString() + "px";
  var calculate_rotation = (23.5 - 3.5 * rotation) / 4

  catImage.style.animationDuration = calculate_rotation.toString() + "s";

  //style.animation = "spin " + rotation.toString() + "s linear infinite";
  //style.animation = "spin 1s linear infinite"
  //width: 240px;
  //height: 120px;
  //margin: -60px 0 0 -120px;
  //-webkit-animation: spin 4s linear infinite;
  //-moz-animation: spin 4s linear infinite;
  //animation: spin 4s linear infinite;
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



    // Set States for Buttons depending on state stream from server
    var receivedAllowedrpcmethods;
    receivedAllowedrpcmethods = message.toObject().allowedrpcmethods;
    controlButtonStates(receivedAllowedrpcmethods);

    // Set Wallet balance
    setWalletbalance(message.toObject().currentWalletbalanceSatoshi, message.toObject().currentWalletbalanceSek);

    // Set Taxi Ride cost
    setTaxiRideCost(message.toObject().currentTaxirideSatoshi, message.toObject().currentTaxiRideSek);

    //Set Average Payment Amount and No of Payments
    setAveragePaymentAmount(
        message.toObject().averagePaymentAmountSatoshi,
        message.toObject().averagePaymentAmountSek,
        message.toObject().avaregeNumberOfPayments);

    // Depending of Accelation, Speed and time(paymentintervall) the cats dimensions and rotation speed will change
    setCatDimensions(message.toObject().accelarationPercent, message.toObject().speedPercent, message.toObject().avaregeNumberOfPayments);


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
