package services

import (
	"context"
	"fmt"

	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/mariajdab/flight-price/internal/providers"
)

type FlightService struct {
	providers []providers.Flight
}

func NewFlightService(providers ...providers.Flight) *FlightService {
	// skyscanner
	// google flights
	// ademus
	return &FlightService{
		providers: providers,
	}
}

func (s *FlightService) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) []entity.FlightSearchResponse {
	for _, provider := range s.providers {
		flights, err := provider.SearchFlights(ctx, criteria)
		if err != nil {
			fmt.Println("en SearchFlights", err)
		}
		fmt.Println(flights)
	}
	return nil
}
