package amadeus

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type ProviderAmadeus struct {
	client *Client
}

func NewProviderAmadeus(client *Client) *ProviderAmadeus {
	return &ProviderAmadeus{client: client}
}

func (a *ProviderAmadeus) SearchFlights(ctx context.Context, criteria entity.SearchReq) ([]entity.Flight, error) {

	return nil, nil
}
