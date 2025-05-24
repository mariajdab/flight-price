package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mariajdab/flight-price/helper"
	"github.com/mariajdab/flight-price/internal/entity"
	services "github.com/mariajdab/flight-price/internal/flights/service"
)

var funcMap = template.FuncMap{
	"subtract": func(a, b int) int { return a - b },
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// PageData holds all data passed to templates
type PageData struct {
	FlightResponse  *entity.FlightPriceResponse
	SearchPerformed bool
	Token           string
	TokenPreview    string // First few characters of token for display
}

type Server struct {
	httpServer *http.Server
	flight     *services.FlightService
}

type jwtCustomClaims struct {
	jwt.RegisteredClaims
}

// Authentication handler - generates a token and sets it as a cookie
func (s *Server) authenticate(c echo.Context) error {
	claims := &jwtCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	// Set token in cookie
	cookie := new(http.Cookie)
	cookie.Name = "jwt_token"
	cookie.Value = t
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)

	// Redirect back to the home page
	return c.Redirect(http.StatusSeeOther, "/public/")
}

// Home page handler - checks for the token cookie
func (s *Server) homePage(c echo.Context) error {
	tokenValue := ""
	cookie, err := c.Cookie("jwt_token")
	if err == nil && cookie.Value != "" {
		tokenValue = cookie.Value
	}
	return c.Render(http.StatusOK, "index.html", PageData{
		Token:        tokenValue,
		TokenPreview: tokenValue,
	})
}

// logout handler - clears the token cookie
func (s *Server) logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "jwt_token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/public/")
}

// Simple check handler (no authentication required)
func (s *Server) simpleCheck(c echo.Context) error {
	return c.String(http.StatusOK, "API is running")
}

// handleFlightSearch - handles the POST request from the flight search form
func (s *Server) handleFlightSearch(c echo.Context) error {
	// Token is valid, process the search request
	origin := c.FormValue("origin")
	destination := c.FormValue("destination")
	date := c.FormValue("date")

	req := entity.FlightSearchParam{
		Origin:        origin,
		Destination:   destination,
		DateDeparture: date,
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			log.Printf("the variable %s is not vaild: %s\n\n", err.Tag(), err.Field())
			return c.NoContent(http.StatusBadRequest)
		}
	}

	orignCode := helper.CityToIATACode(origin)
	destCode := helper.CityToIATACode(destination)

	if orignCode == "" || destCode == "" {
		log.Printf("warning: city not support: %s %s", orignCode, destCode)
		return c.NoContent(http.StatusBadRequest)
	}

	cookie, err := c.Cookie("jwt_token")
	tokenValue := ""
	if err == nil && cookie.Value != "" {
		tokenValue = cookie.Value
	}

	resp := s.flight.SearchFlights(context.Background(), req)

	result := &resp
	if len(result.FlightByProvider) == 0 {
		result = nil
	}

	return c.Render(http.StatusOK, "index.html", PageData{
		FlightResponse:  result,
		SearchPerformed: true,
		Token:           tokenValue,
		TokenPreview:    tokenValue,
	})
}

func New(flightService *services.FlightService, tls *tls.Config) *Server {
	e := echo.New()

	// Set up middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize template renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("assets/templates/*.html")),
	}
	e.Renderer = renderer

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
		flight:     flightService,
	}

	public := e.Group("/public")
	public.GET("/", srv.homePage)
	public.POST("/auth", srv.authenticate)
	public.GET("/logout", srv.logout)
	public.GET("/api/check", srv.simpleCheck) // Simple check endpoint

	private := e.Group("/private")
	private.POST("/flights/search", srv.handleFlightSearch)

	index := e.Group("/")
	index.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusOK, "/public/")
	})

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
