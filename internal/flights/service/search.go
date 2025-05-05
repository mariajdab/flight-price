package services

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/mariajdab/flight-price/internal/providers"
)

type FlightService struct {
	providers []providers.Flight
}

func NewFlightService(providers ...providers.Flight) *FlightService {
	return &FlightService{
		providers: providers,
	}
}

func (s *FlightService) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) []entity.FlightSearchResponse {
	return nil
}
