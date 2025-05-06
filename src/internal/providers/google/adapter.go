package google

import (
	"context"
	"errors"
	"github.com/mariajdab/flight-price/helper"
	"github.com/mariajdab/flight-price/internal/entity"
)

type GoogleFlight struct {
	client *Client
}

func NewAdapterGoogleFlight(client *Client) *GoogleFlight {
	return &GoogleFlight{client: client}
}

func (p *GoogleFlight) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) (entity.FlightSearchResponse, error) {
	origin := helper.CityToGoogleCode(criteria.Origin)
	destination := helper.CityToGoogleCode(criteria.Destination)

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
