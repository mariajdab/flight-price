package google

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type GoogleFlight struct {
	client *Client
}

func NewAdapterGoogleFlight(client *Client) *GoogleFlight {
	return &GoogleFlight{client: client}
}

func (p *GoogleFlight) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
