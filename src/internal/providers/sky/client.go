package sky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mariajdab/flight-price/internal/entity"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

// this client use RAPID API
const providerName = "flights-sky"

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
	itineraries, err := c.getFlightItineraries(ctx, params)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in sky-flght when trying to getFlightItineraries: %w", err)
	}

	resp, err := itineraryPreProcessResponse(itineraries)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in itineraryPreProcessResponse: %w", err)
	}

	return resp, nil
}

func (c *Client) getFlightItineraries(ctx context.Context, params entity.FlightSearchParam) ([]entity.FlightItinerary, error) {
	const flightSearchEndpoint = "flights/search-one-way"

	baseURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, flightSearchEndpoint))
	if err != nil {
		return nil, err
	}

	// building the query parameters
	query := url.Values{}
	query.Set("fromEntityId", params.Origin)
	query.Set("toEntityId", params.Destination)
	query.Set("departDate", params.DateDeparture)
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
	req.Header.Set("x-rapidapi-host", "flights-sky.p.rapidapi.com")
	req.Header.Set("x-rapidapi-key", c.apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get flight itineraries (status %d): %s", resp.StatusCode, string(errorBody))
	}
	var flights entity.FlightSkyResp
	if err := json.NewDecoder(resp.Body).Decode(&flights); err != nil {
		log.Println("internal error during decode response from sky provider", err)
		return nil, err
	}

	return flights.Data.Itineraries, nil
}

func itineraryPreProcessResponse(itineraries []entity.FlightItinerary) (entity.FlightSearchResponse, error) {
	if len(itineraries) == 0 {
		return entity.FlightSearchResponse{}, errors.New("empty offers list")
	}

	// initialize with the first flight
	cheapest := itineraries[0]
	fastest := itineraries[0]
	var fastestDuration = math.MaxInt64

	resp := entity.FlightSearchResponse{
		Flights: make([]entity.Flight, 0, len(itineraries)),
	}

	for _, it := range itineraries {
		// check cheapest flight
		if it.Price.Amount < cheapest.Price.Amount {
			cheapest = it
		}

		// check for prevent panic
		if len(it.Legs) == 0 {
			log.Println("the flight offer does not have itineraries: ", it.ID)
			continue
		}
		if it.Legs[0].Duration < fastestDuration {
			fastest = it
			fastestDuration = it.Legs[0].Duration
		}

		segments := createSegments(it.Legs[0].Segments)

		// save flight data in a useful struct
		resp.Flights = append(resp.Flights, entity.Flight{
			DurationMinutes: it.Legs[0].Duration,
			Segments:        segments,
			Price:           it.Price.Amount,
		})
	}
	resp.Provider = providerName
	resp.Currency = entity.DefaultCurrency
	resp.Cheapest = createFlightFromItinerary(cheapest)
	resp.Fastest = createFlightFromItinerary(fastest)

	return resp, nil
}

func createSegments(segmentsData []entity.SegmentSky) []entity.Segment {
	segments := make([]entity.Segment, 0, len(segmentsData))
	for _, s := range segmentsData {
		segments = append(segments, entity.Segment{
			DepartureAirport:   s.Origin.Name,
			DepartureTime:      formatDate(s.DepartureDate),
			DestinationAirport: s.Destination.Name, // Use DestinationAirportName
			ArrivalTime:        formatDate(s.ArrivalDate),
		})
	}
	return segments
}

// createFlightFromItinerary is a helper function to create Flight from Itinerary
func createFlightFromItinerary(it entity.FlightItinerary) entity.Flight {
	l := it.Legs[0]
	segments := createSegments(l.Segments)

	return entity.Flight{
		DurationMinutes: l.Duration,
		Price:           it.Price.Amount,
		Segments:        segments,
	}
}

func formatDate(dateStr string) string {
	layout := "2006-01-02T15:04:05"
	parsedTime, _ := time.Parse(layout, dateStr)

	formattedTime := parsedTime.Format("2006-01-02 15:04:05")
	return formattedTime
}
