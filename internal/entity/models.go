package entity

import "time"

type Flight struct {
	Airline     string
	Origin      string
	Destination string
	Departure   time.Time
	Arrival     time.Time
	Price       float64
}

type FlightSearchParam struct {
	Origin        string
	Destination   string
	DateDeparture time.Time
	Adults        int32
	Currency      string
}

type FlightSearchResponse struct {
	Provider string
	Currency string
	Cheapest Flight
	Fastest  Flight
	Flights  []Flight
	Error    error
}

type FlightAmadeusResp struct {
	Data []FlightOffer `json:"data"`
}

type FlightOffer struct {
	Itineraries []struct {
		Duration string `json:"duration"`
		Segments []struct {
			Departure struct {
				At time.Time `json:"at"`
			} `json:"departure"`
			Arrival struct {
				At time.Time `json:"at"`
			} `json:"arrival"`
			CarrierCode string `json:"carrierCode"`
		} `json:"segments"`
	} `json:"itineraries"`
	Price struct {
		Total    string `json:"total"`
		Currency string `json:"currency"`
	} `json:"price"`
	Dictionaries []struct {
		Carriers any `json:"carriers"`
	} `json:"dictionaries"`
}
