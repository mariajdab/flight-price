package providers

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type Flight interface {
	SearchFlights(ctx context.Context, criteria entity.FlightSearchParam) ([]entity.Flight, error)
}
