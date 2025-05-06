package amadeus

import (
	"context"
	"encoding/json"
	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetFlights_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/security/oauth2/token" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-token",
				"token_type":   "Bearer",
			})

		case r.URL.Path == "/v2/shopping/flight-offers" && r.Method == http.MethodGet:
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			query := r.URL.Query()
			assert.Equal(t, "JFK", query.Get("originLocationCode"))
			assert.Equal(t, "LAX", query.Get("destinationLocationCode"))

			resp := entity.FlightAmadeusResp{
				Data: []entity.FlightOffer{
					{
						ID: "1",
						Price: struct {
							Total    string `json:"total"`
							Currency string `json:"currency"`
						}{Total: "200"},
						Itineraries: []entity.ItinerariesAmadeus{
							{
								Duration: "PT2H30M",
								Segments: []entity.SegmentAmadeus{
									{
										Departure: struct {
											IataCode string `json:"iataCode"`
											At       string `json:"at"`
										}{IataCode: "JFK", At: "2024-01-01T10:00:00"},
										Arrival: struct {
											IataCode string `json:"iataCode"`
											At       string `json:"at"`
										}{IataCode: "LAX", At: "2024-01-01T12:30:00"},
									},
								},
							},
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

		default:
			t.Errorf("Request inesperada: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer testServer.Close()

	client := NewClient(http.Client{}, entity.Provider{
		BaseURL: testServer.URL,
		Apikey:  "test-api-key",
		Secret:  "test-secret",
		Timeout: time.Second,
	})

	result, err := client.GetFlights(context.Background(), entity.FlightSearchParam{
		Origin:        "JFK",
		Destination:   "LAX",
		DateDeparture: "2024-01-01",
	})

	require.NoError(t, err)
	assert.Equal(t, "Amadeus", result.Provider)
	assert.Equal(t, 200.0, result.Cheapest.Price)
	assert.Contains(t, result.Flights[0].Segments[0].DepartureAirport, "JFK")
}

func TestOffersPreProcessResponse(t *testing.T) {
	offers := []entity.FlightOffer{
		{
			ID: "cheapest",
			Price: struct {
				Total    string `json:"total"`
				Currency string `json:"currency"`
			}{Total: "100.00"},
			Itineraries: []entity.ItinerariesAmadeus{
				{Duration: "PT2H30M", Segments: []entity.SegmentAmadeus{}},
			},
		},
		{
			ID: "fastest",
			Price: struct {
				Total    string `json:"total"`
				Currency string `json:"currency"`
			}{Total: "150.00"},
			Itineraries: []entity.ItinerariesAmadeus{
				{Duration: "PT1H15M", Segments: []entity.SegmentAmadeus{}},
			},
		},
	}

	resp, err := offersPreProcessResponse(offers)
	if err != nil {
		t.Fatalf("offersPreProcessResponse() error = %v", err)
	}

	if resp.Cheapest.Price != 100.00 {
		t.Errorf("Expected cheapest price 100.00, got %v", resp.Cheapest.Price)
	}
	if resp.Fastest.DurationMinutes != 75 {
		t.Errorf("Expected fastest duration 75m, got %v", resp.Fastest.DurationMinutes)
	}
}
