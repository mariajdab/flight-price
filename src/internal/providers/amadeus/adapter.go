package amadeus

import (
	"context"
	"errors"

	"github.com/mariajdab/flight-price/helper"
	"github.com/mariajdab/flight-price/internal/entity"
)

type Amadeus struct {
	client *Client
}

func NewAdapterAmadeus(client *Client) *Amadeus {
	return &Amadeus{client: client}
}

func (p *Amadeus) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	origin := helper.CityToIATACode(criteria.Origin)
	destination := helper.CityToIATACode(criteria.Destination)

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
