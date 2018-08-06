package customer_struct

import (
	"jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"github.com/markdaws/simple-state-machine"
)

type Customer struct {
Title                   string
CustomerStateMachine    *ssm.StateMachine
PaymentStreamStarted    bool
lastRecievedPriceInfo   *taxi_grpc_api.Price
lastRecievedPriceAccept *taxi_grpc_api.AckNackResponse
stateBeforeHaltPayments ssm.State
lastReceivedInvoice     *taxi_grpc_api.PaymentRequest
receivedTaxiInvoiceButNotPaid bool
}