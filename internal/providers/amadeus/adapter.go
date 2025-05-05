package amadeus

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type ProviderAmadeus struct {
	client *Client
}

func NewProviderAmadeus(client *Client) *ProviderAmadeus {
	return &ProviderAmadeus{client: client}
}

func (p *ProviderAmadeus) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
