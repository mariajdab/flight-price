package amadeus

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type Amadeus struct {
	client *Client
}

func NewAdapterAmadeus(client *Client) *Amadeus {
	return &Amadeus{client: client}
}

func (p *Amadeus) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
