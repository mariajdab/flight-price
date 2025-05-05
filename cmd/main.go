package main

import (
	"crypto/tls"
	"github.com/mariajdab/flight-price/api"
	"github.com/mariajdab/flight-price/config"
	"github.com/mariajdab/flight-price/internal/entity"
	services "github.com/mariajdab/flight-price/internal/flights/service"
	"github.com/mariajdab/flight-price/internal/providers/amadeus"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
)

const PROD = "production"

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config variables: ", err)
	}

	tlsConfig := &tls.Config{}
	if c.AppEnv == PROD && c.AppBaseURL != "" {
		// for production use let's encrypt
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("certs"),
			HostPolicy: autocert.HostWhitelist(c.AppBaseURL),
		}
		tlsConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}
	} else { // for development use self certificated
		cert, err := tls.LoadX509KeyPair("cert.pem", "cert.key")
		if err != nil {
			log.Fatalf("Error on cert and key: %v", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	cfg := entity.ProvConfig{
		Providers: []entity.Provider{
			{
				Name:    entity.AmadeusProvider,
				BaseURL: c.AmadeusBaseURL,
				Apikey:  c.AmadeusAPIKey,
				Secret:  c.AmadeusAPISecret,
				Timeout: c.ClientTimeout,
			},
			{
				Name:    entity.SKyRapidProvider,
				BaseURL: c.SkyRapidBaseURL,
				Apikey:  c.GoogleFlightRapidBaseURL,
				Timeout: c.ClientTimeout,
			},
			{
				Name:    entity.GoogleFlightRapidProvider,
				BaseURL: c.GoogleFlightRapidBaseURL,
				Apikey:  c.GoogleFlightRapidAPIKey,
				Timeout: c.ClientTimeout,
			},
		},
	}

	httpClient := http.Client{}

	amadeusClient := amadeus.NewClient(httpClient, cfg.Providers[0])
	amadeusAdapter := amadeus.NewAdapterAmadeus(amadeusClient)

	flightService := services.NewFlightService(
		amadeusAdapter,
	)

	server := api.New(flightService, tlsConfig)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
