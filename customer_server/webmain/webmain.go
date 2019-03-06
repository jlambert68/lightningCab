// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package webmain

import (
	//"crypto/tls"
	//"github.com/jlambert68/lightningCab/customer_gui/frontend/bundle"
	"net/http"
	"path"
	//"strings"
	//"time"

	//"github.com/gorilla/websocket"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	//"github.com/lpar/gzipped"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"jlambert/lightningCab/customer_server/gui_gopherbackend/backend"
	//"jlambert/lightningCab/customer_gui"
	//"github.com/jlambert68/lightningCab/grpc_api/proto/server"
	protoLibrary "jlambert/lightningCab/customer_gui_grpc-web/go/_proto/examplecom/library"
	"os"
	"fmt"
)

var logger *logrus.Logger

func init() {
	/*logger = logrus.StandardLogger()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
		DisableSorting:  true,
	})
	*/
	// Should only be done from init functions
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(logger.Out, logger.Out, logger.Out))
}

func Webmain(customerLogger *logrus.Logger,
	cbTT1 backend.CallBackFunctionType_AskForPrice,
	cbTT2 backend.CallBackFunctionType_AcceptPrice,
	cbTT3 backend.CallBackFunctionType_HaltPayments,
	cbTT4 backend.CallBackFunctionType_LeaveTaxi,
	cbTT5 backend.CallBackFunctionType_PriceAndStateRespons) {

	// Set backend logger to same as Customer logger
	logger = customerLogger
	backend.SetLogger(customerLogger)

	// Set up Callback functions to main Customer Server
	backend.SetAskForPrice(cbTT1)
	backend.SetAcceptPrice(cbTT2)
	backend.SetHaltPayments(cbTT3)
	backend.SetLeaveTaxi(cbTT4)
	backend.SetLPriceAndStateRespons(cbTT5)

	grpcServer := grpc.NewServer()
	protoLibrary.RegisterCustomer_UIServer(grpcServer, &backend.Customer_UI{})
	wrappedServer := grpcweb.WrapServer(grpcServer)

	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}

	//addr := "localhost:10000"
	port := 9091
	/*
	httpsSrv := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
		// Some security settings
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
	}*/
	httpsSrv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}

	logger.Info("Serving on https://" + string(port))
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(dir)
	var appendPath string
	basePath := path.Base(dir)

	switch basePath {
	case "lightningCab":
		appendPath = "./customer_gui/"

	case "customer_server":
		appendPath = "./customer_gui/"

	default:
		appendPath = "./"
	}
	//logger.Fatal(httpsSrv.ListenAndServeTLS("./cert.pem", "./key.pem"))
	logger.Fatal(httpsSrv.ListenAndServeTLS(appendPath + "apache.crt", appendPath + "apache.key"))

}

/*
func folderReader(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") {
			// Use contents of index.html for directory, if present.
			req.URL.Path = path.Join(req.URL.Path, "index.html")
		}
		fn.ServeHTTP(w, req)
	}
}
*/
