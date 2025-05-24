package amadeus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mariajdab/flight-price/internal/entity"
)

const providerName = "Amadeus"

type Client struct {
	httpClient http.Client
	baseURL    string
	apikey     string
	secret     string
	timeout    time.Duration
}

func NewClient(httpClient http.Client, configProvider entity.Provider) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    configProvider.BaseURL,
		apikey:     configProvider.Apikey,
		secret:     configProvider.Secret,
		timeout:    configProvider.Timeout,
	}
}

func (c *Client) GetFlights(ctx context.Context, params entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return entity.FlightSearchResponse{}, err
	}

	offers, err := c.getFlightOffers(ctx, token, params)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in getFlightOffers: %w", err)
	}

	resp, err := offersPreProcessResponse(offers)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error in offersProcessResponse: %w", err)
	}

	return resp, nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	const tokenEndpoint = "v1/security/oauth2/token"

	tokenURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, tokenEndpoint))
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Add("client_id", c.apikey)
	data.Add("client_secret", c.secret)

	req, err := http.NewRequest("POST", tokenURL.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(fmt.Errorf("error creando request: %v", err))
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(fmt.Errorf("error creando request: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token (status %d): %s", resp.StatusCode, string(errorBody))
	}

	var auth struct {
		AccessToken string `json:"access_token"`
		State       string `json:"state"` // TODO: add a check for this
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		log.Println("internal error during decode txn response", err)
		return "", err
	}

	return auth.AccessToken, nil
}

func (c *Client) getFlightOffers(ctx context.Context, token string, params entity.FlightSearchParam) ([]entity.FlightOffer, error) {
	const flightOfferEndpoint = "v2/shopping/flight-offers"

	baseURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, flightOfferEndpoint))
	if err != nil {
		return nil, err
	}

	// building the query parameters
	query := url.Values{}
	query.Set("originLocationCode", params.Origin)
	query.Set("destinationLocationCode", params.Destination)
	query.Set("departureDate", params.DateDeparture)
	query.Set("adults", entity.DefaultAdults)
	query.Set("travelClass", entity.DefaultTravelClass)
	query.Set("currencyCode", entity.DefaultCurrency)

	baseURL.RawQuery = query.Encode()
	flightOffersURL := baseURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, flightOffersURL, nil)
	if err != nil {
		log.Println(fmt.Errorf("error creando request: %v", err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get flight offers (status %d): %s", resp.StatusCode, string(errorBody))
	}

	var flights entity.FlightAmadeusResp

	if err := json.NewDecoder(resp.Body).Decode(&flights); err != nil {
		log.Println("internal error during decode response from amadeus provider", err)
		return nil, err
	}

	return flights.Data, nil
}

// offersPreProcessResponse aim to preprocess the data and obtain the cheapest and fast flight for the provider
func offersPreProcessResponse(offers []entity.FlightOffer) (entity.FlightSearchResponse, error) {
	if len(offers) == 0 {
		return entity.FlightSearchResponse{}, errors.New("empty offers list")
	}

	// initialize with the first flight
	cheapest := offers[0]
	fastest := offers[0]
	lastPriceCheapest := math.MaxFloat32

	fastestDuration, err := parseDuration(fastest.Itineraries[0].Duration)
	if err != nil {
		fmt.Println("Error parsing fastest duration:", err)
		return entity.FlightSearchResponse{}, err
	}

	resp := entity.FlightSearchResponse{
		Flights: make([]entity.Flight, 0, len(offers)),
	}

	for _, offer := range offers {
		price, err := strconv.ParseFloat(offer.Price.Total, 64)
		if err != nil {
			return entity.FlightSearchResponse{}, err
		}

		// check cheapest flight
		if price < lastPriceCheapest {
			cheapest = offer
			lastPriceCheapest = price
		}

		if len(offer.Itineraries) == 0 {
			log.Println("the flight offer does not have itineraries: ", offer)
			continue
		}

		segments := make([]entity.Segment, 0)
		// iterate over itineraries just in case it has more than one item
		for _, it := range offer.Itineraries {
			duration, err := parseDuration(it.Duration)
			if err != nil {
				log.Printf("warning: could not parse duration for offer: %v, error: %v", offer.ID, err)
				continue
			}
			if duration < fastestDuration {
				fastest = offer
				fastestDuration = duration
			}
			for _, s := range it.Segments {
				segments = append(segments, entity.Segment{
					DepartureAirport:   s.Departure.IataCode,
					DepartureTime:      s.Departure.At,
					DestinationAirport: s.Arrival.IataCode,
					ArrivalTime:        s.Arrival.At,
				})
			}
		}

		// save flight data in a useful struct
		resp.Flights = append(resp.Flights, entity.Flight{
			Price:           price,
			DurationMinutes: durationToMinutes(offer.Itineraries[0].Duration),
			Segments:        segments,
		})
	}

	priceCh, err := strconv.ParseFloat(cheapest.Price.Total, 64)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error parsing cheapest price: %w", err)
	}
	priceFt, err := strconv.ParseFloat(fastest.Price.Total, 64)
	if err != nil {
		return entity.FlightSearchResponse{}, fmt.Errorf("error parsing fastest price: %w", err)
	}

	resp.Provider = providerName
	resp.Currency = entity.DefaultCurrency
	resp.Cheapest = createFlightFromOffer(cheapest, priceCh)
	resp.Fastest = createFlightFromOffer(fastest, priceFt)

	return resp, nil
}

func parseDuration(durationStr string) (time.Duration, error) {
	durationStr = strings.TrimPrefix(durationStr, "PT")
	durationStr = strings.Replace(durationStr, "H", "h", 1)
	durationStr = strings.Replace(durationStr, "M", "m", 1)
	return time.ParseDuration(durationStr)
}

// createFlightFromOffer is a helper function to create Flight from Offer
func createFlightFromOffer(offer entity.FlightOffer, price float64) entity.Flight {
	segments := make([]entity.Segment, 0, len(offer.Itineraries[0].Segments))

	for _, s := range offer.Itineraries[0].Segments {

		segments = append(segments, entity.Segment{
			DepartureAirport:   s.Departure.IataCode,
			DepartureTime:      s.Departure.At,
			DestinationAirport: s.Arrival.IataCode,
			ArrivalTime:        s.Arrival.At,
		})
	}

	return entity.Flight{
		ProviderName:    providerName,
		Price:           price,
		DurationMinutes: durationToMinutes(offer.Itineraries[0].Duration),
		Segments:        segments,
	}
}

func durationToMinutes(duration string) int {
	re := regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?$`)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 0 {
		log.Printf("warning: could not parse duration: %v", duration)
		return 0
	}

	var hours int
	if matches[1] != "" {
		h, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("warning: could not parse duration: %v", duration)
			return 0
		}
		hours = h
	}

	var minutes int
	if matches[2] != "" {
		m, err := strconv.Atoi(matches[2])
		if err != nil {
			return 0
		}
		minutes = m
	}

	totalMinutes := (hours * 60) + minutes
	return totalMinutes
}
