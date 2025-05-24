package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/mariajdab/flight-price/internal/entity"
)

// this client use RAPID API
const providerName = "google"

type Client struct {
	httpClient http.Client
	baseURL    string
	apikey     string
	timeout    time.Duration
}

func NewClient(httpClient http.Client, configProvider entity.Provider) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    configProvider.BaseURL,
		apikey:     configProvider.Apikey,
		timeout:    configProvider.Timeout,
	}
}

func (c *Client) GetFlights(ctx context.Context, params entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := c.getTopFlights(ctx, params)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in getTopFlights: %w", err)
	}

	resp, err := flightsPreProcess(flights)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in offersProcessResponse: %w", err)
	}

	return resp, nil
}

func (c *Client) getTopFlights(ctx context.Context, params entity.FlightSearchParam) ([]entity.OtherFlight, error) {
	const (
		flightSearchEndpoint = "flights/search-one-way"
		host                 = "google-flights4.p.rapidapi.com"
	)

	baseURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, flightSearchEndpoint))
	if err != nil {
		return nil, err
	}

	// building the query parameters
	query := url.Values{}
	query.Set("departureId", params.Origin)
	query.Set("arrivalId", params.Destination)
	query.Set("departureDate", params.DateDeparture)

	baseURL.RawQuery = query.Encode()
	flightOffersURL := baseURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, flightOffersURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-host", host)
	req.Header.Set("x-rapidapi-key", c.apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get flights (status %d): %s", resp.StatusCode, string(errorBody))
	}

	var flights entity.FlightGoogleResp

	if err := json.NewDecoder(resp.Body).Decode(&flights); err != nil {
		log.Println("internal error during decode response from amadeus provider", err)
		return nil, err
	}

	return flights.Data.OtherFlights, nil
}

func flightsPreProcess(flights []entity.OtherFlight) (entity.FlightSearchResponse, error) {
	if len(flights) == 0 {
		return entity.FlightSearchResponse{}, errors.New("empty flights list from google-flights")
	}

	// initialize with the first flight
	cheapest := flights[0]
	fastest := flights[0]
	var fastestDuration = math.MaxInt64

	resp := entity.FlightSearchResponse{
		Flights: make([]entity.Flight, 0, len(flights)),
	}

	for _, f := range flights {
		// check cheapest flight
		if f.Price < cheapest.Price {
			cheapest = f
		}

		if f.Duration < fastestDuration {
			fastest = f
			fastestDuration = f.Duration
		}

		segments := createSegments(f.Segments)

		// save flight data in a useful struct
		resp.Flights = append(resp.Flights, entity.Flight{
			Price:           f.Price,
			DurationMinutes: f.Duration,
			Segments:        segments,
		})
	}

	resp.Provider = providerName
	resp.Currency = entity.DefaultCurrency
	resp.Cheapest = createFlightFromOtherFlight(cheapest)
	resp.Fastest = createFlightFromOtherFlight(fastest)

	return resp, nil
}

func createSegments(segmentsData []entity.SegmentGoogleF) []entity.Segment {
	segments := make([]entity.Segment, 0, len(segmentsData))
	for _, s := range segmentsData {
		departureTime, err := formatDate(s.DepartureTime, s.DepartureDate)
		if err != nil {
			log.Println("error in formatting departure time", err, departureTime)
			departureTime = s.DepartureDate // not info abut time only date
		}

		arrivalTime, err := formatDate(s.ArrivalTime, s.ArrivalDate)
		if err != nil {
			log.Println("error in formatting arrival time", err, departureTime)
			arrivalTime = s.ArrivalTime // not info abut time only date
		}

		segments = append(segments, entity.Segment{
			DepartureAirport:   s.DepartureAirportName,
			DepartureTime:      departureTime,
			DestinationAirport: s.ArrivalAirportName, // Use DestinationAirportName
			ArrivalTime:        arrivalTime,
		})
	}
	return segments
}

func createFlightFromOtherFlight(tf entity.OtherFlight) entity.Flight {
	segments := createSegments(tf.Segments)

	return entity.Flight{
		ProviderName:    providerName,
		Price:           tf.Price,
		DurationMinutes: tf.Duration,
		Segments:        segments,
	}
}

func formatDate(timeStr, dateStr string) (string, error) {
	dateTimeStr := fmt.Sprintf("%sT%s:00", dateStr, timeStr)
	layout := "2006-01-02T15:04:05"

	parsedTime, err := time.Parse(layout, dateTimeStr)
	if err != nil {
		return "", err
	}

	customFormat := parsedTime.Format("2006-01-02 15:04:05")
	return customFormat, nil
}
