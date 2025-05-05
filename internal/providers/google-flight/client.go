package sky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mariajdab/flight-price/internal/entity"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

// this client use RAPID API
const providerName = "google-flight"

type Client struct {
	httpClient http.Client
	baseURL    string
	apikey     string
	timeout    time.Duration
}

func NewClient(httpClient http.Client, baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		apikey:     apiKey,
		timeout:    timeout,
	}
}

func (c *Client) getFlightOffers(ctx context.Context, params entity.FlightSearchParam) ([]entity.TopFlights, error) {
	const flightSearchEndpoint = "/flights/search-one-way"

	baseURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, flightSearchEndpoint))
	if err != nil {
		return nil, err
	}

	date := params.DateDeparture.Format(time.DateOnly)

	// building the query parameters
	query := url.Values{}
	query.Set("departureId", params.Origin)
	query.Set("arrivalId", params.Destination)
	query.Set("departureDate", date)
	query.Set("adults", entity.DefaultAdults)
	query.Set("cabinClass", entity.DefaultTravelClass)
	query.Set("currency", entity.DefaultCurrency)

	baseURL.RawQuery = query.Encode()
	flightOffersURL := baseURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, flightOffersURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-host", "google-flights4.p.rapidapi.com")
	req.Header.Set("x-rapidapi-key", c.apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get flight offers from amadeus: %s", resp.Body)
	}

	var flights entity.FlightGoogleResp

	if err := json.NewDecoder(resp.Body).Decode(&flights); err != nil {
		log.Println("internal error during decode response from amadeus provider", err)
		return nil, err
	}

	return flights.Data, nil
}

func flightsPreProcessResponse(flights []entity.TopFlights) (entity.FlightSearchResponse, error) {
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
	resp.Cheapest = createFlightFromTopOffers(cheapest)
	resp.Fastest = createFlightFromTopOffers(fastest)

	return resp, nil
}

func createSegments(segmentsData []entity.SegmentGoogleF) []entity.Segment {
	segments := make([]entity.Segment, 0, len(segmentsData))
	for _, s := range segmentsData {
		departureTime, err := formatDate(s.DepartureTime, s.DepartureDate)
		if err != nil {
			// log here
			departureTime = "2006-01-02 15:04:05"
		}

		arrivalTime, err := formatDate(s.ArrivalTime, s.ArrivalDate)
		if err != nil {
			// log here
			arrivalTime = "2006-01-02 15:04:05"
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

func createFlightFromTopOffers(tf entity.TopFlights) entity.Flight {
	segments := createSegments(tf.Segments)

	return entity.Flight{
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
		fmt.Println("Error parsing time:", err)
		return "", err
	}

	customFormat := parsedTime.Format("2006-01-02 15:04:05")
	return customFormat, nil
}
