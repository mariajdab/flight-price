package sky

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type SkyRapid struct {
	client *Client
}

func NewAdapterSkyRapid(client *Client) *SkyRapid {
	return &SkyRapid{client: client}
}

func (p *SkyRapid) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
