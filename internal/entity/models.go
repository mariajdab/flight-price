package entity

import "time"

type Flight struct {
	Airline      string
	FlightNumber string
	Origin       string
	Destination  string
	Departure    time.Time
	Arrival      time.Time
	Price        float64
	Currency     string
}

type SearchReq struct {
	Origin      string
	Destination string
	Date        time.Time
}

type FlightSearchResponse struct {
	Provider string
	Flights  []Flight
	Error    error
}
