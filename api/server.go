package api

import (
	"net/http"
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
func searchFlightsHandler(c echo.Context) error {
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

	// TODO: implement flight search logic
	// for now, just return a dummy response
	flights := []Flight{
		{Origin: origin, Destination: destination, Date: date},
	}

	return c.JSON(http.StatusOK, flights)
}

func NewServer() *echo.Echo {
	e := echo.New()

	// define the routes and groups
	public := e.Group("/public/v1")
	public.GET("/generate-token", generateTokenHandler)

	private := e.Group("/private/v1")
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte("secret"),
	}

	private.Use(echojwt.WithConfig(config))
	private.GET("/flights/search", searchFlightsHandler)

	return e
}
