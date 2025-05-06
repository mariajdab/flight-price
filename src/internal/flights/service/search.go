package services

import (
	"context"
	"fmt"
	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/mariajdab/flight-price/internal/providers"
	"math"
)

type FlightService struct {
	providers []providers.Flight
}

func NewFlightService(providers ...providers.Flight) *FlightService {
	return &FlightService{
		providers: providers,
	}
}

func (s *FlightService) SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) entity.FlightPriceResponse {
	allCheapest := make([]entity.Flight, 0, len(s.providers))
	allFastest := make([]entity.Flight, 0, len(s.providers))
	allProviderFlights := make([]entity.FlightSearchResponse, 0, len(s.providers))

	for _, provider := range s.providers {
		resp, err := provider.SearchFlights(ctx, criteria)
		if err != nil {
			fmt.Println("en SearchFlights", err)
			return entity.FlightPriceResponse{}
		}

		allCheapest = append(allCheapest, resp.Cheapest)
		allFastest = append(allFastest, resp.Fastest)
		allProviderFlights = append(allProviderFlights, resp)
	}

	cheapest := getGlobalCheapestOrFastestFlight(allCheapest, "cheapest")
	fastest := getGlobalCheapestOrFastestFlight(allFastest, "fastest")

	return entity.FlightPriceResponse{
		Cheapest:         cheapest,
		Fastest:          fastest,
		FlightByProvider: allProviderFlights,
	}
}

func getGlobalCheapestOrFastestFlight(flights []entity.Flight, criteria string) entity.Flight {
	bestFlight := flights[0]
	var fastestDuration = math.MaxInt64

	for _, f := range flights {
		if criteria == "cheapest" {
			if f.Price < bestFlight.Price {
				bestFlight = f
			}
		} else if criteria == "fastest" {
			if f.DurationMinutes < fastestDuration {
				bestFlight = f
				fastestDuration = f.DurationMinutes
			}
		}
	}
	return bestFlight
}
