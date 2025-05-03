package sky_rapid

import (
	"context"
	"github.com/mariajdab/flight-price/internal/entity"
)

type ProviderSkyRapid struct {
	client *Client
}

func NewProviderAmadeus(client *Client) *ProviderSkyRapid {
	return &ProviderSkyRapid{client: client}
}

func (a *ProviderSkyRapid) SearchFlights(ctx context.Context, criteria entity.SearchReq) ([]entity.Flight, error) {

	return nil, nil
}
