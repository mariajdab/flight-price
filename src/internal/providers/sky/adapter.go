package sky

import (
	"context"
	"errors"

	"github.com/mariajdab/flight-price/helper"
	"github.com/mariajdab/flight-price/internal/entity"
)

type SkyRapid struct {
	client *Client
}

func NewAdapterSkyRapid(client *Client) *SkyRapid {
	return &SkyRapid{client: client}
}

func (p *SkyRapid) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	origin := helper.CityToSkyCode(criteria.Origin)
	destination := helper.CityToSkyCode(criteria.Destination)

	if origin == "" || destination == "" {
		return entity.FlightSearchResponse{}, errors.New("origin or destination not supported")
	}

	criteria.Origin = origin
	criteria.Destination = destination

	flights, err := p.client.GetFlights(ctx, criteria)

	if err != nil {
		return entity.FlightSearchResponse{}, err
	}
	return flights, nil
}
