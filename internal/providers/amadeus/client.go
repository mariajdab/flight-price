package amadeus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mariajdab/flight-price/internal/entity"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTravelClass = "Economy"
	defaultCurrency    = "USD"
)

type Client struct {
	httpClient http.Client
	baseURL    string
	apikey     string
	secret     string
	timeout    time.Duration
}

func NewClient(httpClient http.Client, baseURL, apiKey, secret string, timeout time.Duration) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		apikey:     apiKey,
		secret:     secret,
		timeout:    timeout,
	}
}

func (c *Client) GetFlights(ctx context.Context, params entity.FlightSearchParam) error {

}

func (c *Client) getAccessToken() (string, error) {
	const tokenEndpoint = "v1/security/oauth2/token"

	bodyData := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.apikey,
		"client_secret": c.secret,
	}

	tokenURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, tokenEndpoint))
	if err != nil {
		return "", err
	}

	jsonBody, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %s", resp.Body)
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

	date := params.DateDeparture.Format(time.DateOnly)

	// building the query parameters
	query := url.Values{}
	query.Set("originLocationCode", params.Origin)
	query.Set("destinationLocationCode", params.Destination)
	query.Set("departureDate", date)
	query.Set("adults", fmt.Sprintf("%v", params.Adults))
	query.Set("travelClass", defaultTravelClass)
	query.Set("currencyCode", defaultCurrency)

	baseURL.RawQuery = query.Encode()
	flightOffersURL := baseURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, flightOffersURL, nil)
	if err != nil {
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
		return nil, fmt.Errorf("failed to get flight offers from amadeus: %s", resp.Body)
	}

	var flights entity.FlightAmadeusResp

	if err := json.NewDecoder(resp.Body).Decode(&flights); err != nil {
		log.Println("internal error during decode response from amadeus provider", err)
		return nil, err
	}

	return flights.Data, nil
}
