package backends

type PublishInvoiceSettled func(invoice string) //, eventSrv *eventsource.Server)

type Backend interface {
	Connect() error

	// Amount in satoshis and expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, err error)

	// Pay an array of payment requests
	CompletePaymentRequests(paymentRequests []string, awaitResponse bool) (error)
}
