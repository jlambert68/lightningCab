package lightningServer

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	lndrpc "github.com/lightningnetwork/lnd/lnrpc"
	lndmacaroons "github.com/lightningnetwork/lnd/macaroons"
	googleGRPc "google.golang.org/grpc"
	googleCred "google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	//"gopkg.in/macaroon.v1"
	"io/ioutil"
	"os/user"
	"path"
)

func LndServer() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Cannot get current user:", err)
		return
	}
	fmt.Println("The user home directory: " + usr.HomeDir)
	tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := googleCred.NewClientTLSFromFile(tlsCertPath, "")
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

	opts := []googleGRPc.DialOption{
		googleGRPc.WithTransportCredentials(tlsCreds),
		googleGRPc.WithBlock(),
		googleGRPc.WithPerRPCCredentials(lndmacaroons.NewMacaroonCredential(mac)),
	}

	conn, err := googleGRPc.Dial("localhost:10009", opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return
	}
	client := lndrpc.NewLightningClient(conn)

	ctx := context.Background()
	getInfoResp, err := client.GetInfo(ctx, &lndrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println("Cannot get info from node:", err)
		return
	}
	spew.Dump(getInfoResp)
}
