package sky

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type ProviderSkyRapid struct {
	client *Client
}

func NewProviderSkyRapid(client *Client) *ProviderSkyRapid {
	return &ProviderSkyRapid{client: client}
}

func (p *ProviderSkyRapid) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
