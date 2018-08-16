package backends

import (
	//"github.com/donovanhide/eventsource"
	"github.com/lightningnetwork/lnd/lnrpc"
	"io"
	"errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"encoding/hex"
	"golang.org/x/net/context"
	"fmt"
	"time"
	"jlambert/lightningCab/taxi_server/taxi_grpc_api"


)

type LND struct {
	GRPCHost     string `long:"grpchost" Description:"Host of the gRPC interface of LND"`
	CertFile     string `long:"certfile" Description:"TLS certificate for the LND gRPC and REST services"`
	MacaroonFile string `long:"macaroonfile" Description:"Admin macaroon file for authentication. Set to an empty string for no macaroon"`

	client lnrpc.LightningClient
	ctx    context.Context
}

func (lnd *LND) Connect() error {
	log.Info("Entering Connect() in lnd.go for Customer")
	creds, err := credentials.NewClientTLSFromFile(lnd.CertFile, "")

	if err != nil {
		log.Error("Failed to read certificate for LND gRPC")

		return err
	}

	con, err := grpc.Dial(lnd.GRPCHost, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Error("Failed to connect to LND gRPC server")

		return err
	}

	lnd.ctx = context.Background()

	if lnd.MacaroonFile != "" {
		macaroon, err := getMacaroon(lnd.MacaroonFile)

		if macaroon == nil && err != nil {
			log.Error("Failed to read admin macaroon file of LND")
		}

		lnd.ctx = metadata.NewOutgoingContext(lnd.ctx, macaroon)
	}

	lnd.client = lnrpc.NewLightningClient(con)

	return err
}

func getMacaroon(macaroonFile string) (macaroon metadata.MD, err error) {
	data, err := ioutil.ReadFile(macaroonFile)

	if err == nil {
		macaroon = metadata.Pairs("macaroon", hex.EncodeToString(data))
	}

	return macaroon, err
}


func (lnd *LND) GetInvoice(message string, amount int64, expiry int64) (invoice string, err error) {
	var response *lnrpc.AddInvoiceResponse

	if message != "" {
		response, err = lnd.client.AddInvoice(lnd.ctx, &lnrpc.Invoice{
			Memo:   message,
			Value:  amount,
			Expiry: expiry,
		})

	} else {
		response, err = lnd.client.AddInvoice(lnd.ctx, &lnrpc.Invoice{
			Value:  amount,
			Expiry: expiry,
		})
	}

	if err != nil {
		return "", err
	}

	return response.PaymentRequest, err
}

func (lnd *LND) SubscribeInvoices(callback PublishInvoiceSettled) error { //, eventSrv *eventsource.Server) error {
	stream, err := lnd.client.SubscribeInvoices(lnd.ctx, &lnrpc.InvoiceSubscription{})

	if err != nil {
		return err
	}

	wait := make(chan struct{})

	go func() {
		for {
			invoice, streamErr := stream.Recv()

			if streamErr == io.EOF {
				err = errors.New("lost connection to LND gRPC")

				close(wait)

				return
			}

			if streamErr != nil {
				err = streamErr

				close(wait)

				return
			}

			if invoice.Settled {
				callback(invoice.PaymentRequest) // , eventSrv)
			}

		}

	}()

	<-wait
	//dd
	return err
}

func (lnd *LND) CompletePaymentRequests(paymentRequests []*taxi_grpc_api.PaymentRequest, awaitResponse bool) (err error) {

	log.Info(("Entering 'CompletePaymentRequests'"))

	ctx, cancel := context.WithCancel(lnd.ctx)
	defer cancel()

	payStream, err := lnd.client.SendPayment(ctx)
	if err != nil {
		return err
	}

	for _, payReqData := range paymentRequests {
		payReq := payReqData.LightningPaymentRequest
		sendReq := &lnrpc.SendRequest{PaymentRequest: payReq}
		err := payStream.Send(sendReq)
		if err != nil {
			return err
		}
	}

	if awaitResponse {
		for range paymentRequests {
			resp, err := payStream.Recv()
			if err != nil {
				return err
			}
			if resp.PaymentError != "" {
				return fmt.Errorf("received payment error: %v",
					resp.PaymentError)
			}
		}
	} else {
		// We are not waiting for feedback in the form of a response, but we
		// should still wait long enough for the server to receive and handle
		// the send before cancelling the request.
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

/*
func  (lnd *LND) RetrieveGetInfo() {
	ctx := context.Background()
	getInfoResp, err := lnd.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
		//getInfoResp, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println("Cannot get info from node:", err)
		return
	}
	(getInfoResp)
}
*/

func (lnd *LND) GetWalletBalance() (*lnrpc.WalletBalanceResponse, error){
	//ctxb := context.Background()
	//ctx, _ := context.WithTimeout(context.Background(), 60)

	req := &lnrpc.WalletBalanceRequest{}
	resp, err := lnd.client.WalletBalance(lnd.ctx, req)

	return resp, err
}

func (lnd *LND) GetWalletChannelBalance() (*lnrpc.ChannelBalanceResponse, error){

	req := &lnrpc.ChannelBalanceRequest{}
	resp, err := lnd.client.ChannelBalance(lnd.ctx, req)

	return resp, err
}