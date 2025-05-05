package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mariajdab/flight-price/api"
	"github.com/mariajdab/flight-price/config"
	"golang.org/x/crypto/acme/autocert"
)

const PROD = "production"

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config variables: ", err)
	}

	server := &http.Server{
		Addr:         ":" + os.Getenv("SERVER_PORT"),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if c.AppEnv == PROD && c.AppBaseURL != "" {
		// for production use let's encrypt
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("certs"),
			HostPolicy: autocert.HostWhitelist(c.AppBaseURL),
		}
		server.TLSConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}
	} else { // for development use self certificated
		cert, err := tls.LoadX509KeyPair("cert.pem", "cert.key")
		if err != nil {
			log.Fatalf("Error on cert and key: %v", err)
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	handler := api.NewServer()

	server.Handler = handler

	go func() {
		log.Println("Starting server on :8443")
		if err := server.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting HTTPS server: %v", err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutting down the server: %v", err)
	}
	log.Println("Shutdown server")
}
