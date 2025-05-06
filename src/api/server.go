package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/mariajdab/flight-price/internal/flights/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// Flight represents a flight
type Flight struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
}

type Server struct {
	httpServer *http.Server
	flightSvc  *services.FlightService
}

type jwtCustomClaims struct {
	jwt.RegisteredClaims
}

func generateTokenHandler(c echo.Context) error {
	claims := &jwtCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, t)
}

// searchFlightsHandler handles the GET /flights/search request
func (s *Server) handleFlightSearch(c echo.Context) error {
	// authenticate the request using JWT
	user := c.Get("user")
	token := user.(*jwt.Token)
	_, ok := token.Claims.(*jwtCustomClaims)
	if !ok || !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	// if authenticated, proceed with the request
	origin := c.QueryParam("origin")
	destination := c.QueryParam("destination")
	date := c.QueryParam("date")

	flights := s.flightSvc.SearchFlights(
		context.Background(),
		entity.FlightSearchParam{
			Origin:        origin,
			Destination:   destination,
			DateDeparture: date,
		})

	return c.JSON(http.StatusOK, flights)
}

func New(flightSvc *services.FlightService, tls *tls.Config) *Server {
	e := echo.New()

	public := e.Group("/public/v1")
	public.GET("/generate-token", generateTokenHandler)
	public.File("/", "./assets/index.html")
	private := e.Group("/private/v1")
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte("secret"),
	}

	server := &http.Server{
		Addr:         ":8443",
		Handler:      e,
		TLSConfig:    tls,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	srv := &Server{
		httpServer: server,
		flightSvc:  flightSvc,
	}

	private.Use(echojwt.WithConfig(config))
	private.GET("/flights/search", srv.handleFlightSearch)

	return srv
}

func (s *Server) Start() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on :8443")
		if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting HTTPS server: %v", err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Shutdown server")
	return nil
}
