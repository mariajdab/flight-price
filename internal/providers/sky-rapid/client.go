package sky_rapid

import (
	"net/http"
	"time"
)

type Client struct {
	httpClient http.Client
}

func NewClient(httpClient http.Client, baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		httpClient: httpClient,
	}
}
