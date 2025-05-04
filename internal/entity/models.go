package entity

import "time"

type Flight struct {
	Airline         string
	OriginCode      string
	DestinationCode string
	DepartureTime   time.Time
	ArrivalTime     time.Time
	Price           float64
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
	ID          string `json:"id"`
	Itineraries []struct {
		Duration string `json:"duration"`
		Segments []struct {
			Departure struct {
				IataCode string    `json:"iataCode"`
				At       time.Time `json:"at"`
			} `json:"departure"`
			Arrival struct {
				IataCode string    `json:"iataCode"`
				At       time.Time `json:"at"`
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
