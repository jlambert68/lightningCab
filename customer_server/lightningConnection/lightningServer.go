package lightningConnection

import (
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"

	"time"
	"os"
	//"github.com/op/go-logging"
	//"github.com/donovanhide/eventsource"
	"strconv"
	"crypto/sha256"
	"encoding/hex"
	//"jlambert/lightningCab/customer_server/lightningConnection/backends"

	"jlambert/lightningCab/taxi_server/taxi_grpc_api"
	"github.com/sirupsen/logrus"
	"jlambert/lightningCab/customer_server/lightningConnection/backends"
)

var lndClient lnrpc.LightningClient

const eventChannel = "invoiceSettled"

const couldNotParseError = "Could not parse values from request"

//var eventSrv *eventsource.Server

var pendingInvoices []PendingInvoice

type PendingInvoice struct {
	Invoice string
	Hash    string
	Expiry  time.Time
}

// To use the pendingInvoice type as event for the EventSource stream
func (pending PendingInvoice) Id() string    { return "" }
func (pending PendingInvoice) Event() string { return "" }
func (pending PendingInvoice) Data() string  { return pending.Hash }

type invoiceRequest struct {
	Amount  int64
	Message string
}

type invoiceResponse struct {
	Invoice string
	Expiry  int64
}

type invoiceSettledRequest struct {
	InvoiceHash string
}

type invoiceSettledResponse struct {
	Settled bool
}

type errorResponse struct {
	Error string
}

type TaxiPaysToll func(check bool) (err error) //, eventSrv *eventsource.Server)

type BackendLayer1 interface {
	Connect() error
}

var callbackToToll TaxiPaysToll

/*
func InitLndServerConnection() {

	initLog()

	initConfig()

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Cannot get current user:", err)
		return
	}
	fmt.Println("The user home directory: " + usr.HomeDir)
	tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return
	}

	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Cannot read macaroon file", err)
		return
	}


	mac := &macaroon.Macaroon{}
	//mac := &lndmacarons.MacaroonCredential{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		fmt.Println("Cannot unmarshal macaroon", err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(lndmacarons.NewMacaroonCredential(mac)),
	}

	conn, err := grpc.Dial("localhost:10009", opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return
	}
	lndClient = lnrpc.NewLightningClient(conn)
}
*/

func LigtningMainService(customerLogger *logrus.Logger) { //cbTT TaxiPaysToll) {
	//callbackToToll = cbTT

	var err error

	initLog(customerLogger)
	backends.UseLogger(customerLogger)

	initConfig()

	err = backend.Connect()

	if err == nil {

		log.Debug("Starting ticker to clear expired invoices")

		// A bit longer than the expiry time to make sure the invoice doesn't show as settled if it isn't (affects just invoiceSettledHandler)
		duration := time.Duration(cfg.TipExpiry + 10)
		ticker := time.Tick(duration * time.Second)

		go func() {
			for {
				select {
				case <-ticker:
					log.Debug("Start checking for expired Invoices : ")
					now := time.Now()

					for index := len(pendingInvoices) - 1; index >= 0; index-- {
						invoice := pendingInvoices[index]

						if now.Sub(invoice.Expiry) > 0 {
							log.Debug("Invoice expired: " + invoice.Invoice)

							pendingInvoices = append(pendingInvoices[:index], pendingInvoices[index+1:]...)
						}

					}

				}

			}

		}()

		log.Info("Subscribing to invoices")

		go func() {
			err = cfg.LND.SubscribeInvoices(publishInvoiceSettled) //, eventSrv)

			if err != nil {
				log.Error("Failed to subscribe to invoices: " + fmt.Sprint(err))

				os.Exit(1)
			}

		}()

		select {}
	}

}

// Callback for backends
func publishInvoiceSettled(invoice string) { //, eventSrv *eventsource.Server) {
	for index, pending := range pendingInvoices {
		if pending.Invoice == invoice {
			log.Info("Invoice settled: " + invoice)

			//eventSrv.Publish([]string{eventChannel}, pending)

			pendingInvoices = append(pendingInvoices[:index], pendingInvoices[index+1:]...)

			//Callback to previous layer to inform that Invoice is paid
			_ = callbackToToll(false)

			break
		}

	}

}

/*
func invoiceSettledHandler(writer http.ResponseWriter, request *http.Request) {
	errorMessage := couldNotParseError

	if request.Method == http.MethodPost {
		var body invoiceSettledRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.InvoiceHash != "" {
				settled := true

				for _, pending := range pendingInvoices {
					if pending.Hash == body.InvoiceHash {
						settled = false

						break
					}

				}

				writer.Write(marshalJson(invoiceSettledResponse{
					Settled: settled,
				}))

				return

			}

		}

	}

	log.Error(errorMessage)

	writeError(writer, errorMessage)
}
*/

func PayReceivedInvoicesFromTaxi(invoices []*taxi_grpc_api.PaymentRequest) (err error) {
	err = nil

	err = backend.CompletePaymentRequests(invoices, true)

	return err
}

func CreateInvoice(message string, amount int64, expire int64) (newInvoice PendingInvoice, err error) {

	var errorMessage string = "Amount is zero"

	if amount != 0 {
		invoice, err := backend.GetInvoice(message, amount, expire)

		if err == nil {
			logMessage := "Created invoice with amount of " + strconv.FormatInt(amount, 10) + " satoshis"

			if message != "" {
				logMessage += " with message \"" + message + "\""
			}

			sha := sha256.New()
			sha.Write([]byte(invoice))

			hash := hex.EncodeToString(sha.Sum(nil))

			expiryDuration := time.Duration(expire) * time.Second

			log.Debug(logMessage)

			newInvoice = PendingInvoice{
				Invoice: invoice,
				Hash:    hash,
				Expiry:  time.Now().Add(expiryDuration),
			}

			pendingInvoices = append(pendingInvoices, newInvoice)

			return newInvoice, err

		} else {
			errorMessage = "Failed to create invoice"
			newInvoice = PendingInvoice{"", "", time.Now()}
		}

	}

	log.Error(errorMessage)
	return newInvoice, err
}

/*
func getInvoiceHandler(writer http.ResponseWriter, request *http.Request) {
	errorMessage := couldNotParseError

	if request.Method == http.MethodPost {
		var body invoiceRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.Amount != 0 {
				invoice, err := backend.GetInvoice(body.Message, body.Amount, cfg.TipExpiry)

				if err == nil {
					logMessage := "Created invoice with amount of " + strconv.FormatInt(body.Amount, 10) + " satoshis"

					if body.Message != "" {
						logMessage += " with message \"" + body.Message + "\""
					}

					sha := sha256.New()
					sha.Write([]byte(invoice))

					hash := hex.EncodeToString(sha.Sum(nil))

					expiryDuration := time.Duration(cfg.TipExpiry) * time.Second

					log.Info(logMessage)

					pendingInvoices = append(pendingInvoices, PendingInvoice{
						Invoice: invoice,
						Hash:    hash,
						Expiry:  time.Now().Add(expiryDuration),
					})

					writer.Write(marshalJson(invoiceResponse{
						Invoice: invoice,
						Expiry:  cfg.TipExpiry,
					}))

					return

				} else {
					errorMessage = "Failed to create invoice"
				}

			}

		}

	}

	log.Error(errorMessage)

	writeError(writer, errorMessage)
}
*/

/*
func notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	writeError(writer, "Not found")
}

func writeError(writer http.ResponseWriter, message string) {
	writer.WriteHeader(http.StatusBadRequest)

	writer.Write(marshalJson(errorResponse{
		Error: message,
	}))
}

func marshalJson(data interface{}) []byte {
	response, _ := json.MarshalIndent(data, "", "    ")

	return response
}

*/

/*
func  (lnd *LND) RetrieveGetInfo() {
	ctx := context.Background()
	a,b : = cfg.LND.
	getInfoResp, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println("Cannot get info from node:", err)
		return
	}
	spew.Dump(getInfoResp)
}
*/



/*
func LndServer() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Cannot get current user:", err)
		return
	}
	fmt.Println("The user home directory: " + usr.HomeDir)
	tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return
	}

	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Cannot read macaroon file", err)
		return
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		fmt.Println("Cannot unmarshal macaroon", err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(lndmacarons.NewMacaroonCredential(mac)),
	}

	conn, err := grpc.Dial("localhost:10009", opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return
	}
	client := lnrpc.NewLightningClient(conn)

	ctx := context.Background()
	getInfoResp, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println("Cannot get info from node:", err)
		return
	}
	spew.Dump(getInfoResp)
}
*/

func CustomerWalletbalance() (*lnrpc.WalletBalanceResponse, error) {
	rest, err := cfg.LND.GetWalletBalance()
		return rest, err
}

func CustomerChannelbalance() (*lnrpc.ChannelBalanceResponse, error) {
	rest, err := cfg.LND.GetWalletChannelBalance()
	return rest, err

}
