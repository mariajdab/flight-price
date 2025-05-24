package services

import (
	"context"
	"log"
	"sync"

	"github.com/mariajdab/flight-price/internal/entity"
	"github.com/mariajdab/flight-price/internal/providers"
)

const (
	criteriaCheapest = "cheapest"
	criteriaFastest  = "fastest"
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
	var wg sync.WaitGroup

	allCheapest := make([]entity.Flight, 0, len(s.providers))
	allFastest := make([]entity.Flight, 0, len(s.providers))
	allProviderFlights := make([]entity.FlightSearchResponse, 0, len(s.providers))

	resultChan := make(chan struct {
		resp         entity.FlightSearchResponse
		providerName string
		err          error
	}, len(s.providers))

	for _, provider := range s.providers {
		wg.Add(1)
		go func(p providers.Flight) {
			defer wg.Done()
			resp, err := p.SearchFlights(ctx, criteria)
			resultChan <- struct {
				resp         entity.FlightSearchResponse
				providerName string
				err          error
			}{resp, resp.Provider, err}
		}(provider)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.err != nil {
			log.Printf("provider %s get error in SearchFlights: %v", result.providerName, result.err)
			continue
		}

		allCheapest = append(allCheapest, result.resp.Cheapest)
		allFastest = append(allFastest, result.resp.Fastest)
		allProviderFlights = append(allProviderFlights, result.resp)
	}

	if len(allCheapest) == 0 {
		return entity.FlightPriceResponse{}
	}

	cheapest := getGlobalBestFlight(allCheapest, criteriaCheapest)
	fastest := getGlobalBestFlight(allFastest, criteriaFastest)

	return entity.FlightPriceResponse{
		OriginName:       criteria.Origin,
		DestinationName:  criteria.Destination,
		Cheapest:         cheapest,
		Fastest:          fastest,
		FlightByProvider: allProviderFlights,
	}
}

func getGlobalBestFlight(flights []entity.Flight, criteria string) entity.Flight {
	if len(flights) == 0 {
		return entity.Flight{}
	}

	bestFlight := flights[0]

	switch criteria {
	case criteriaCheapest:
		for _, f := range flights[1:] {
			if f.Price < bestFlight.Price {
				bestFlight = f
			}
		}
	case criteriaFastest:
		bestDuration := bestFlight.DurationMinutes
		for _, f := range flights[1:] {
			if f.DurationMinutes < bestDuration {
				bestFlight = f
				bestDuration = f.DurationMinutes
			}
		}
	default:
		log.Printf("invalid criteria: %s", criteria)
	}

	return bestFlight
}
