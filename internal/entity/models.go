package entity

import "time"

const (
	DefaultTravelClass = "Economy"
	DefaultCurrency    = "USD"
	DefaultAdults      = "1"
)

type Flight struct {
	Price           float64
	DurationMinutes int       `json:"total_duration_minutes"`
	Segments        []Segment `json:"segments"`
}

type Segment struct {
	DepartureAirport   string `json:"departureAirport"`
	DestinationAirport string `json:"destinationAirport"`
	DepartureTime      string
	ArrivalTime        string
}

type FlightSearchParam struct {
	Origin        string
	Destination   string
	DateDeparture time.Time
}

type FlightSearchResponse struct {
	OriginName      string
	DestinationName string
	Provider        string
	Currency        string
	Cheapest        Flight
	Fastest         Flight
	Flights         []Flight
	Error           error
}

type FlightAmadeusResp struct {
	Data []FlightOffer `json:"data"`
}

type FlightSkyResp struct {
	Data []FlightItinerary `json:"itineraries"`
}

type FlightGoogleResp struct {
	Data []TopFlights
}

// FlightOffer represent amadeus response of a flight search
type FlightOffer struct {
	ID          string               `json:"id"`
	Itineraries []ItinerariesAmadeus `json:"itineraries"`
	Price       struct {
		Total    string `json:"total"`
		Currency string `json:"currency"`
	} `json:"price"`
}

// FlightItinerary represent flights-sky response of a flight search
type FlightItinerary struct {
	ID    string `json:"id"`
	Price struct {
		Amount float64 `json:"raw"`
	} `json:"price"`
	Legs []struct {
		Segments  []SegmentSky `json:"segments"`
		Duration  int          `json:"durationInMinutes"`
		Departure time.Time    `json:"departure"`
		Arrival   time.Time    `json:"arrival"`
	} `json:"legs"`
}

// TopFlights represent flights google-flight response of a flight search
type TopFlights struct {
	Price    float64          `json:"price"`
	Duration int              `json:"duration"`
	Segments []SegmentGoogleF `json:"segments"`
	Stops    int              `json:"stops"`
}

type ItinerariesAmadeus struct {
	Duration string           `json:"duration"`
	Segments []SegmentAmadeus `json:"segments"`
}

type SegmentAmadeus struct {
	Departure struct {
		IataCode string `json:"iataCode"`
		At       string `json:"at"`
	} `json:"departure"`
	Arrival struct {
		IataCode string `json:"iataCode"`
		At       string `json:"at"`
	} `json:"arrival"`
	CarrierCode string `json:"carrierCode"`
}

type SegmentGoogleF struct {
	DepartureAirportName string `json:"departureAirportName"`
	ArrivalAirportName   string `json:"arrivalAirportName"`
	DepartureDate        string `json:"departureDate"`
	ArrivalDate          string `json:"arrivalDate"`
	DepartureTime        string `json:"departureTime"`
	ArrivalTime          string `json:"arrivalTime"`
}

type SegmentSky struct {
	Origin struct {
		Name string `json:"name"`
	} `json:"origin"`
	Destination struct {
		Name string `json:"name"`
	} `json:"destination"`
	DepartureDate string `json:"departure"`
	ArrivalDate   string `json:"arrival"`
}
