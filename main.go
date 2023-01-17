package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

var (
	parameters                           WhSvrParameters
	webhookNamespace, webhookServiceName, webhookLabelSelector string
)

func init() {
	// webhook server running namespace
	webhookNamespace = os.Getenv("POD_NAMESPACE")
}

func main() {
	// init command flags
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&webhookServiceName, "service-name", "sidecar-injector", "Webhook service name.")
	flag.StringVar(&parameters.sidecarCfgFile, "sidecar-config-file", "/etc/webhook/config/sidecarconfig.yaml", "Sidecar injector configuration file.")
	flag.StringVar(&webhookLabelSelector, "label-selector", "sidecar-injector=enabled", "Namespace label selector for the mutating webhook.")
	// flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	// flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	dnsNames := []string{
		webhookServiceName,
		webhookServiceName + "." + webhookNamespace,
		webhookServiceName + "." + webhookNamespace + ".svc",
	}
	commonName := webhookServiceName + "." + webhookNamespace + ".svc"

	org := "was.net.pl"
	caPEM, certPEM, certKeyPEM, err := generateCert([]string{org}, dnsNames, commonName)
	if err != nil {
		glog.Fatalf("Failed to generate ca and certificate key pair: %v", err)
	}

	pair, err := tls.X509KeyPair(certPEM.Bytes(), certKeyPEM.Bytes())
	if err != nil {
		glog.Fatalf("Failed to load certificate key pair: %v", err)
	}
	/*
		pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
		if err != nil {
			glog.Errorf("Failed to load key pair: %v", err)
		}
	*/

	sidecarConfig, err := loadConfig(parameters.sidecarCfgFile)
	if err != nil {
		glog.Fatalf("Failed to load configuration: %v", err)
	}

	// create or update the mutatingwebhookconfiguration
	err = createOrUpdateMutatingWebhookConfiguration(caPEM, webhookServiceName, webhookNamespace, webhookLabelSelector)
	if err != nil {
		glog.Fatalf("Failed to create or update the mutating webhook configuration: %v", err)
	}

	whsvr := &WebhookServer{
		sidecarConfig: sidecarConfig,
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc(webhookInjectPath, whsvr.serve)
	whsvr.server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			glog.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	glog.Info("Server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.server.Shutdown(context.Background())
}
