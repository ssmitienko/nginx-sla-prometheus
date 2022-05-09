package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type srvHTTP struct {
	handler http.Handler
}

var (
	lstPtr *string
	bckPtr *string
	usrPtr *string
	pwdPtr *string
	pfxPtr *string
)

func (srv *srvHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.handler.ServeHTTP(w, r)
}

func newSRV(handler http.Handler) *srvHTTP {
	return &srvHTTP{handler}
}

func waitAndShutdown(srv *http.Server) {

	var wait = time.Second * 15

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Println("Shutdown failed:", err)
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func main() {
	lstPtr = flag.String("listen", "0.0.0.0:9043", "address:port to listen to")
	bckPtr = flag.String("backend", "http://localhost/nginx-sla/", "Nginx sla backend")
	pfxPtr = flag.String("prefix", "nginx_sla", "Metric name prefix")
	usrPtr = flag.String("user", "", "Username for basic auth")
	pwdPtr = flag.String("password", "", "Password for basic auth")

	tlsPtr := flag.Bool("tls", false, "enable TLS")
	tlsCrt := flag.String("cert", "server.crt", "TLS certificate")
	tlsKey := flag.String("key", "server.key", "TLS key")

	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/metrics", handleMetrics).Methods("GET")

	srv := newSRV(handlers.LoggingHandler(os.Stdout, r))

	if *tlsPtr == true {
		cert, err := tls.LoadX509KeyPair(*tlsCrt, *tlsKey)
		if err != nil {
			log.Println(err)
			return
		}

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			Certificates:             []tls.Certificate{cert},
			CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.CurveP384, tls.CurveP521},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			},
		}
		srv := &http.Server{
			Addr:         *lstPtr,
			Handler:      srv,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		go func() { log.Fatal(srv.ListenAndServeTLS(*tlsCrt, *tlsKey)) }()
		waitAndShutdown(srv)
	} else {
		srv := &http.Server{
			Addr:    *lstPtr,
			Handler: srv,
		}
		go func() { log.Fatal(srv.ListenAndServe()) }()
		waitAndShutdown(srv)
	}
}
