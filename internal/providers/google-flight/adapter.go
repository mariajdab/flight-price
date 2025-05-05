package google_flight

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type ProviderGoogleFlight struct {
	client *Client
}

func NewProviderGoogleFlight(client *Client) *ProviderGoogleFlight {
	return &ProviderGoogleFlight{client: client}
}

func (p *ProviderGoogleFlight) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
