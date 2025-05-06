package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetFlights_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/flights/search-one-way", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "JFK", query.Get("departureId"))
		assert.Equal(t, "LAX", query.Get("arrivalId"))
		assert.Equal(t, "2024-01-01", query.Get("departureDate"))

		assert.Equal(t, "google-flights4.p.rapidapi.com", r.Header.Get("x-rapidapi-host"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-rapidapi-key"))

		resp := entity.FlightGoogleResp{
			Data: entity.DataGoogle{
				OtherFlights: []entity.OtherFlight{
					{
						Price:    200,
						Duration: 180,
						Segments: []entity.SegmentGoogleF{
							{
								DepartureTime:        "10:00",
								DepartureDate:        "2024-01-01",
								ArrivalTime:          "12:30",
								ArrivalDate:          "2024-01-01",
								DepartureAirportName: "JFK",
								ArrivalAirportName:   "LAX",
							},
						},
					},
					{
						Price:    300,
						Duration: 150,
						Segments: []entity.SegmentGoogleF{
							{
								DepartureTime:        "08:00",
								DepartureDate:        "2024-01-01",
								ArrivalTime:          "10:30",
								ArrivalDate:          "2024-01-01",
								DepartureAirportName: "JFK",
								ArrivalAirportName:   "LAX",
							},
						},
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer testServer.Close()

	client := NewClient(http.Client{}, entity.Provider{
		BaseURL: testServer.URL,
		Apikey:  "test-api-key",
		Timeout: time.Second,
	})

	result, err := client.GetFlights(context.Background(), entity.FlightSearchParam{
		Origin:        "JFK",
		Destination:   "LAX",
		DateDeparture: "2024-01-01",
	})

	require.NoError(t, err)
	assert.Equal(t, "google", result.Provider)
	assert.Len(t, result.Flights, 2)
	assert.Equal(t, float64(200), result.Cheapest.Price)
	assert.Equal(t, 150, result.Fastest.DurationMinutes)
	assert.Equal(t, "2024-01-01 08:00:00", result.Fastest.Segments[0].DepartureTime)
}

func TestClient_GetFlights_HTTPError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer testServer.Close()

	client := NewClient(http.Client{}, entity.Provider{
		BaseURL: testServer.URL,
		Apikey:  "test-api-key",
		Timeout: time.Second,
	})

	_, err := client.GetFlights(context.Background(), entity.FlightSearchParam{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestFlightsPreProcess_EmptyList(t *testing.T) {
	_, err := flightsPreProcess([]entity.OtherFlight{})
	require.Error(t, err)
	assert.Equal(t, "empty flights list from google-flights", err.Error())
}

func TestFlightsPreProcess_CheapestAndFastest(t *testing.T) {
	flights := []entity.OtherFlight{
		{Price: 500, Duration: 200},
		{Price: 300, Duration: 150},
		{Price: 400, Duration: 100},
	}

	resp, err := flightsPreProcess(flights)
	require.NoError(t, err)
	assert.Equal(t, float64(300), resp.Cheapest.Price)
	assert.Equal(t, 100, resp.Fastest.DurationMinutes)
}

func TestCreateSegments_TimeFormatting(t *testing.T) {
	segment := entity.SegmentGoogleF{
		DepartureTime:        "invalid-time",
		DepartureDate:        "2024-01-01",
		ArrivalTime:          "15:00",
		ArrivalDate:          "2024-01-01",
		DepartureAirportName: "JFK",
		ArrivalAirportName:   "LAX",
	}

	segments := createSegments([]entity.SegmentGoogleF{segment})
	require.Len(t, segments, 1)
	assert.Equal(t, "2024-01-01", segments[0].DepartureTime)
	assert.Equal(t, "2024-01-01 15:00:00", segments[0].ArrivalTime)
}
