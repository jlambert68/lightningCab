package backends

import (
	"jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type PublishInvoiceSettled func(invoice string) //, eventSrv *eventsource.Server)

type Backend interface {
	Connect() error

	// Amount in satoshis and expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, err error)

	// Pay an array of payment requests
	CompletePaymentRequests(paymentRequests []*taxi_grpc_api.PaymentRequest, awaitResponse bool) (err error)

	//Get Wallet Balance
	GetWalletBalance() (*lnrpc.WalletBalanceResponse, error)

	//Get Channel Balance
	GetWalletChannelBalance() (*lnrpc.ChannelBalanceResponse, error)
}
